package pomodoro

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Category constants
const (
	CategoryPomodoro = "Pomodoro"
	CategoryShortBreak = "ShortBreak"
	CategoryLongBreak = "LongBreak"
)

// State constants
const (
	StateNotStarted = iota
	StateRunning
	StatePaused
	StateDone
	StateCancelled
)

type Interval struct {
	ID int64
	StartTime time.Time
	PlannedDuration time.Duration
	ActualDuration time.Duration
	Category string
	State int
}

// for implementing the repository pattern,
// this interface will be used to abstract the data source
type Repository interface {
	Create(i Interval) (int64, error) // create an interval
	Update(i Interval) error // update an interval
	ByID(id int64) (Interval, error) // retrieve an interval by id
	Last() (Interval, error) // find the last interval
	Breaks(n int) ([]Interval, error) // retrieve a given number of Interval items of type break (short or long)
}

var (
	ErrNoIntervals = errors.New("No intervals")
	ErrIntervalNotRunning = errors.New("Interval not running")
	ErrIntervalCompleted = errors.New("Interval is completed or cancelled")
	ErrInvalidState = errors.New("Invalid State")
	ErrInvalidID = errors.New("Invalid ID")
)

type IntervalConfig struct {
	repo Repository
	PomodoroDuration time.Duration
	ShortBreakDuration time.Duration
	LongBreakDuration time.Duration
}

type Callback func(Interval)

func NewConfig(repo Repository, pomodoro, shortBreak, longBreak time.Duration) *IntervalConfig {
	c := &IntervalConfig{
		repo: repo,
		PomodoroDuration: 25 * time.Minute,
		ShortBreakDuration: 5 * time.Minute,
		LongBreakDuration: 15 * time.Minute,
	}

	if pomodoro > 0 {
		c.PomodoroDuration = pomodoro
	}
	if shortBreak > 0 {
		c.ShortBreakDuration = shortBreak
	}
	if longBreak > 0 {
		c.LongBreakDuration = longBreak
	}

	return c
}

// determines the next interval category based on the last interval from the repository according to pomodoro technique rules
// after each Pomodoro interval, there’s a short break, and after four pomodoros, there’s a long break
func nextCategory(r Repository) (string, error) {
	// retrieves the last interval from repo
	li, err := r.Last()

	// if no last interval (eg, for the first execution)
	if err != nil && err == ErrNoIntervals {
		return CategoryPomodoro, nil
	}
	if err != nil {
		return "", err
	}
	
	// if last interval was break, then next catergory will be pomodoro interval
	// since breaks are followed by a Pomodoro interval accd to rules
	if li.Category == CategoryLongBreak || li.Category == CategoryShortBreak {
		return CategoryPomodoro, nil
	}

	// hereafter, it has been determined that last interval was NOT a break
	// i.e; last interval was a pomodoro interval

	//  fetches the last three intervals categorized as breaks (short or long)
	lastBreaks, err := r.Breaks(3)
	if err != nil {
		return "", err
	}

	// if there are fewer than 3 breaks, next break will be short
	if len(lastBreaks) < 3 {
		return CategoryShortBreak, nil
	}

	// iterates over the last three breaks, looking for a long break
	// if one is found, then current break will be short
	for _, i := range lastBreaks {
		if i.Category == CategoryLongBreak {
			return CategoryShortBreak, nil
		}
	}

	// if none of the last 3 breaks were long
	// since after every four Pomodoros, a long break is taken accd to rules
	return CategoryLongBreak, nil
}

func tick(ctx context.Context, id int64, config *IntervalConfig, start, periodic, end Callback) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	i, err := config.repo.ByID(id)
	if err != nil {
		return err
	}

	expire := time.After(i.PlannedDuration - i.ActualDuration)

	start(i)

	for {
		select {
		case <- ticker.C:
			i, err := config.repo.ByID(id)
			if err != nil {
				return err
			}

			if i.State == StatePaused {
				return nil
			}

			i.ActualDuration += time.Second
			if err := config.repo.Update(i); err != nil {
				return err
			}
			periodic(i)

		case <- expire:
			i, err := config.repo.ByID(id)
			if err != nil {
				return err
			}

			i.State = StateDone
			end(i)
			return config.repo.Update(i)
		case <- ctx.Done():
			i, err := config.repo.ByID(id)
			if err != nil {
				return err
			}

			i.State = StateCancelled
			return config.repo.Update(i)
		}
	}
}

func newInterval(config *IntervalConfig) (Interval, error) {
	i := Interval{}
	category, err := nextCategory(config.repo)
	if err != nil {
		return i, err
	}

	i.Category = category

	switch category {
	case CategoryPomodoro:
		i.PlannedDuration = config.PomodoroDuration
	case CategoryShortBreak:
		i.PlannedDuration = config.ShortBreakDuration
	case CategoryLongBreak:
		i.PlannedDuration = config.LongBreakDuration
	}

	if i.ID, err = config.repo.Create(i); err != nil {
		return i, err
	}

	return i, nil
}

func GetInterval(config *IntervalConfig) (Interval, error) {
	// i := Interval{}
	var err error

	i, err := config.repo.Last()

	if err != nil && err != ErrNoIntervals {
		return i, err
	}

	// if last interval is active
	if err == nil && i.State != StateCancelled && i.State != StateDone {
		return i, nil
	}

	// if the last interval is inactive or unavailable
	return newInterval(config)
}

func(i Interval) Start(ctx context.Context, config *IntervalConfig, start, periodic, end Callback) error {
	switch i.State {
	case StateRunning:
		return nil
	case StateNotStarted:
		i.StartTime = time.Now()
		fallthrough
	case StatePaused:
		i.State = StateRunning
		if err := config.repo.Update(i); err != nil {
			return err
		}
		return tick(ctx, i.ID, config, start, periodic, end)
	case StateCancelled, StateDone:
		return fmt.Errorf("%w: Cannot start", ErrIntervalCompleted)
	default:
		return fmt.Errorf("%w: %d", ErrInvalidState, i.State)
	}
}

func(i Interval) Pause(config *IntervalConfig) error {
	if i.State != StateRunning {
		return ErrIntervalNotRunning
	}
	i.State = StatePaused

	return config.repo.Update(i)
}