package scheduler

import (
	"errors"
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"

	"github.com/robin-vidal/neetrack/internal/catalog"
)

func TestRate(t *testing.T) {
	tests := []struct {
		name string
		o    Outcome
		want fsrs.Rating
	}{
		{
			name: "saw solution always again",
			o:    Outcome{Difficulty: catalog.Easy, Duration: time.Second, SawSolution: true},
			want: fsrs.Again,
		},
		{
			name: "saw solution beats hint",
			o:    Outcome{Difficulty: catalog.Easy, SawSolution: true, UsedHint: true},
			want: fsrs.Again,
		},
		{
			name: "used hint is hard",
			o:    Outcome{Difficulty: catalog.Medium, Duration: time.Minute, UsedHint: true},
			want: fsrs.Hard,
		},
		{
			name: "over twice target is hard",
			o:    Outcome{Difficulty: catalog.Easy, Duration: 2*TargetEasy + time.Minute},
			want: fsrs.Hard,
		},
		{
			name: "over target is good",
			o:    Outcome{Difficulty: catalog.Medium, Duration: TargetMedium + time.Minute},
			want: fsrs.Good,
		},
		{
			name: "at most half target is easy",
			o:    Outcome{Difficulty: catalog.Hard, Duration: TargetHard / 2},
			want: fsrs.Easy,
		},
		{
			name: "between half and full target is good",
			o:    Outcome{Difficulty: catalog.Medium, Duration: TargetMedium - time.Minute},
			want: fsrs.Good,
		},
		{
			name: "exactly at target is good",
			o:    Outcome{Difficulty: catalog.Medium, Duration: TargetMedium},
			want: fsrs.Good,
		},
		{
			name: "exactly at twice target is good not hard",
			o:    Outcome{Difficulty: catalog.Medium, Duration: 2 * TargetMedium},
			want: fsrs.Good,
		},
		{
			name: "unknown difficulty falls back to medium target",
			o:    Outcome{Difficulty: catalog.Difficulty("unknown"), Duration: TargetMedium / 2},
			want: fsrs.Easy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Rate(tt.o); got != tt.want {
				t.Errorf("Rate(%+v) = %v, want %v", tt.o, got, tt.want)
			}
		})
	}
}

func TestDeriveRating(t *testing.T) {
	easy := fsrs.Easy
	tooLow := fsrs.Rating(0)
	tooHigh := fsrs.Rating(5)

	tests := []struct {
		name         string
		o            Outcome
		clientRating *fsrs.Rating
		want         fsrs.Rating
		wantErr      error
	}{
		{
			name: "duration present derives rating",
			o:    Outcome{Difficulty: catalog.Easy, Duration: TargetEasy / 2},
			want: fsrs.Easy,
		},
		{
			name:         "duration absent uses client rating",
			o:            Outcome{Difficulty: catalog.Easy},
			clientRating: &easy,
			want:         fsrs.Easy,
		},
		{
			name:    "duration absent and no client rating errors",
			o:       Outcome{Difficulty: catalog.Easy},
			wantErr: ErrRatingRequired,
		},
		{
			name:         "client rating too low errors",
			o:            Outcome{Difficulty: catalog.Easy},
			clientRating: &tooLow,
			wantErr:      ErrInvalidRating,
		},
		{
			name:         "client rating too high errors",
			o:            Outcome{Difficulty: catalog.Easy},
			clientRating: &tooHigh,
			wantErr:      ErrInvalidRating,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeriveRating(tt.o, tt.clientRating)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("DeriveRating() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("DeriveRating() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("DeriveRating() = %v, want %v", got, tt.want)
			}
		})
	}
}
