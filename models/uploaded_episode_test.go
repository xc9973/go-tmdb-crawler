package models

import (
	"testing"
)

func TestUploadedEpisode_Validate(t *testing.T) {
	tests := []struct {
		name    string
		episode *UploadedEpisode
		wantErr bool
	}{
		{
			name: "valid episode",
			episode: &UploadedEpisode{
				EpisodeID: 1,
				Uploaded:  true,
			},
			wantErr: false,
		},
		{
			name: "missing episode_id",
			episode: &UploadedEpisode{
				Uploaded: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.episode.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
