package engine

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// AnalysisResult holds the engine's output
type AnalysisResult struct {
	BestMove     string   `json:"best_move"`
	Evaluation   float64  `json:"evaluation"`
	Continuation []string `json:"continuation"`
	Depth        int      `json:"depth"`
	MateIn       int      `json:"mate_in"` // 0 = no mate, positive = side to move mates in N, negative = side to move gets mated
	IsMate       bool     `json:"is_mate"`
}

// EngineService manages a pool of Stockfish processes
type EngineService struct {
	binaryPath string
	pool       chan *engineWorker
	mu         sync.Mutex
	workers    []*engineWorker
}

// Status information about the engine pool
type EngineStatus struct {
	ActiveWorkers int  `json:"active_workers"`
	TotalWorkers  int  `json:"total_workers"`
	IsResponsive  bool `json:"is_responsive"`
}

type engineWorker struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	reader *bufio.Reader
}

// NewEngineService creates a new engine service with a pool of workers
func NewEngineService(binaryPath string, poolSize int) (*EngineService, error) {
	svc := &EngineService{
		binaryPath: binaryPath,
		pool:       make(chan *engineWorker, poolSize),
		workers:    make([]*engineWorker, 0, poolSize),
	}

	for i := 0; i < poolSize; i++ {
		worker, err := svc.startWorker()
		if err != nil {
			svc.Stop()
			return nil, fmt.Errorf("failed to start worker %d: %v", i, err)
		}
		svc.workers = append(svc.workers, worker)
		svc.pool <- worker
	}

	return svc, nil
}

func (s *EngineService) startWorker() (*engineWorker, error) {
	cmd := exec.Command(s.binaryPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	worker := &engineWorker{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		reader: bufio.NewReader(stdout),
	}

	// Initialize UCI
	fmt.Fprintln(worker.stdin, "uci")
	for {
		line, err := worker.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(line) == "uciok" {
			break
		}
	}

	// Set multi-threaded engine if possible for speed
	fmt.Fprintln(worker.stdin, "setoption name Threads value 2")
	fmt.Fprintln(worker.stdin, "setoption name Hash value 128")
	fmt.Fprintln(worker.stdin, "isready")
	for {
		line, err := worker.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(line) == "readyok" {
			break
		}
	}

	return worker, nil
}

// PlayOptions holds the parameters for a computer move
type PlayOptions struct {
	Level    int // 1-6
	Elo      int // Manual Elo if Level is 0
	Depth    int // Manual Depth if Level is 0
	Skill    int // Manual Skill if Level is 0
	Movetime int // Maximum time to think
}

// AnalyzePosition evaluates a position and returns the best move and evaluation
func (s *EngineService) AnalyzePosition(ctx context.Context, fen string, depth int, movetime int) (*AnalysisResult, error) {
	// Default depth if invalid
	if depth < 1 {
		depth = 12
	}
	if depth > 30 {
		depth = 30
	}
	if movetime < 500 {
		movetime = 1000
	}
	if movetime > 10000 {
		movetime = 10000
	}

	// Acquire a worker from the pool
	select {
	case worker := <-s.pool:
		defer func() { s.pool <- worker }() // Return worker to pool
		return s.analyze(ctx, worker, fen, depth, movetime)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *EngineService) analyze(ctx context.Context, w *engineWorker, fen string, depth int, movetime int) (*AnalysisResult, error) {
	// Sync engine state to prevent stale results
	fmt.Fprintln(w.stdin, "isready")
	for {
		line, err := w.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(line) == "readyok" {
			break
		}
	}

	fmt.Fprintf(w.stdin, "position fen %s\n", fen)
	if depth > 0 {
		fmt.Fprintf(w.stdin, "go depth %d movetime %d\n", depth, movetime)
	} else {
		fmt.Fprintf(w.stdin, "go movetime %d\n", movetime)
	}

	var result AnalysisResult
	done := make(chan bool)
	errChan := make(chan error)

	// Professional UCI parser — captures depth, mate, cp, pv
	go func() {
		for {
			line, err := w.reader.ReadString('\n')
			if err != nil {
				errChan <- err
				return
			}
			line = strings.TrimSpace(line)

			// Skip lowerbound/upperbound lines — they are incomplete
			if strings.Contains(line, "lowerbound") || strings.Contains(line, "upperbound") {
				continue
			}

			if strings.HasPrefix(line, "info") && strings.Contains(line, "score") {
				parts := strings.Fields(line)
				for i, p := range parts {
					if p == "depth" && i+1 < len(parts) {
						d, _ := strconv.Atoi(parts[i+1])
						result.Depth = d
					}
					if p == "cp" && i+1 < len(parts) {
						score, _ := strconv.ParseFloat(parts[i+1], 64)
						result.Evaluation = score / 100.0
						result.IsMate = false
						result.MateIn = 0
					}
					if p == "mate" && i+1 < len(parts) {
						mateIn, _ := strconv.Atoi(parts[i+1])
						result.MateIn = mateIn
						result.IsMate = true
						if mateIn > 0 {
							result.Evaluation = 99.0
						} else {
							result.Evaluation = -99.0
						}
					}
					if p == "pv" && i+1 < len(parts) {
						limit := i + 8 // capture up to 7 future moves
						if limit > len(parts) {
							limit = len(parts)
						}
						result.Continuation = parts[i+1 : limit]
					}
				}
			}

			if strings.HasPrefix(line, "bestmove") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					result.BestMove = parts[1]
					done <- true
					return
				}
			}
		}
	}()

	select {
	case <-done:
		return &result, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		// Stop engine if request timed out or cancelled
		fmt.Fprintln(w.stdin, "stop")
		return nil, ctx.Err()
	}
}

// PlayMove gets the best move based on Level or individual settings
func (s *EngineService) PlayMove(ctx context.Context, fen string, opts PlayOptions) (*AnalysisResult, error) {
	// Fallback to defaults if Level is not 1-6
	if opts.Level < 1 && opts.Elo == 0 {
		opts.Level = 3 // Default to medium
	}

	// Acquire a worker from the pool
	select {
	case worker := <-s.pool:
		defer func() { s.pool <- worker }()
		return s.play(ctx, worker, fen, opts)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *EngineService) play(ctx context.Context, w *engineWorker, fen string, opts PlayOptions) (*AnalysisResult, error) {
	// Sync engine state
	fmt.Fprintln(w.stdin, "isready")
	for {
		line, err := w.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(line) == "readyok" {
			break
		}
	}

	// Apply strength settings
	skill := 20
	elo := 1200
	limitStrength := false
	depth := 10

	if opts.Level >= 1 && opts.Level <= 6 {
		// Map Levels to professional settings
		switch opts.Level {
		case 1:
			skill, elo, limitStrength, depth = 5, 800, true, 5
		case 2:
			skill, elo, limitStrength, depth = 10, 1200, true, 8
		case 3:
			skill, elo, limitStrength, depth = 15, 1600, true, 10
		case 4:
			skill, elo, limitStrength, depth = 20, 2000, true, 12
		case 5:
			skill, elo, limitStrength, depth = 20, 2500, false, 15
		case 6:
			skill, elo, limitStrength, depth = 20, 3200, false, 20
		}
	} else {
		// Manual ELO mode
		skill = 20
		elo = opts.Elo
		limitStrength = true
		depth = opts.Depth
		if depth == 0 {
			depth = 12 // Default for manual Elo
		}
	}

	fmt.Fprintf(w.stdin, "setoption name Skill Level value %d\n", skill)
	if limitStrength {
		fmt.Fprintln(w.stdin, "setoption name UCI_LimitStrength value true")
		fmt.Fprintf(w.stdin, "setoption name UCI_Elo value %d\n", elo)
	} else {
		fmt.Fprintln(w.stdin, "setoption name UCI_LimitStrength value false")
	}

	fmt.Fprintln(w.stdin, "isready")
	for {
		line, err := w.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(line) == "readyok" {
			break
		}
	}

	fmt.Fprintf(w.stdin, "position fen %s\n", fen)
	if opts.Movetime > 0 {
		fmt.Fprintf(w.stdin, "go depth %d movetime %d\n", depth, opts.Movetime)
	} else {
		fmt.Fprintf(w.stdin, "go depth %d\n", depth)
	}

	var result AnalysisResult
	done := make(chan bool)
	errChan := make(chan error)

	go func() {
		for {
			line, err := w.reader.ReadString('\n')
			if err != nil {
				errChan <- err
				return
			}
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "bestmove") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					result.BestMove = parts[1]
					done <- true
					return
				}
			}
		}
	}()

	var finalErr error
	select {
	case <-done:
		// Move received
	case err := <-errChan:
		finalErr = err
	case <-ctx.Done():
		fmt.Fprintln(w.stdin, "stop")
		finalErr = ctx.Err()
	}

	// ALWAYS reset strength limit and skill level so analysis workers stay full strength
	fmt.Fprintln(w.stdin, "setoption name Skill Level value 20")
	fmt.Fprintln(w.stdin, "setoption name UCI_LimitStrength value false")
	fmt.Fprintln(w.stdin, "isready")
	for {
		line, err := w.reader.ReadString('\n')
		if err != nil {
			if finalErr == nil {
				finalErr = err
			}
			break
		}
		if strings.TrimSpace(line) == "readyok" {
			break
		}
	}

	if finalErr != nil {
		return nil, finalErr
	}

	return &result, nil
}

// Stop terminates all engine processes
func (s *EngineService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, w := range s.workers {
		fmt.Fprintln(w.stdin, "quit")
		// Give it a moment to quit gracefully, then kill
		w.cmd.Process.Kill()
	}
	s.workers = nil
}

// GetStatus checks the health of the engine pool
func (s *EngineService) GetStatus() EngineStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	status := EngineStatus{
		ActiveWorkers: len(s.pool),
		TotalWorkers:  len(s.workers),
		IsResponsive:  false,
	}

	if len(s.workers) == 0 {
		return status
	}

	// Try to ping one worker to check responsiveness
	// We use a short timeout to avoid blocking the admin UI
	select {
	case w := <-s.pool:
		defer func() { s.pool <- w }()
		fmt.Fprintln(w.stdin, "isready")

		timeout := time.After(500 * time.Millisecond)
		done := make(chan bool)
		go func() {
			for {
				line, err := w.reader.ReadString('\n')
				if err != nil {
					return
				}
				if strings.TrimSpace(line) == "readyok" {
					done <- true
					return
				}
			}
		}()

		select {
		case <-done:
			status.IsResponsive = true
		case <-timeout:
			status.IsResponsive = false
		}
	default:
		// All workers are busy, but they are active
		status.IsResponsive = true
	}

	return status
}
