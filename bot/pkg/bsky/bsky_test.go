package bsky

import (
	"testing"
)

func TestGenerateBskyUrl(t *testing.T) {
	post := Post{
		URI: "at://did:plc:tixombvfeipi656t6fpzyi2h/app.bsky.feed.post/3la5fcokczp2e",
		Author: map[string]interface{}{
			"handle": "testhandle.bsky.social",
		},
	}

	url, err := generateBskyUrl(post)
	if err != nil {
		t.Errorf("expected no error generating bsky url: %v", err)
	}

	expectedUrl := "https://bsky.app/profile/testhandle.bsky.social/post/3la5fcokczp2e"
	if url != expectedUrl {
		t.Errorf("expected %s, got %s", expectedUrl, url)
	}

}

func TestExtractRKey(t *testing.T) {
	tests := []struct {
		uri           string
		expectedRkey  string
		testShouldErr bool
	}{
		{"at://did:plc:tixombvfeipi656t6fpzyi2h/app.bsky.feed.post/3la5fcokczp2e", "3la5fcokczp2e", false},
		{"invalid/uri/fmt", "", true},
	}

	for _, tt := range tests {
		rkey, err := extractRKey(tt.uri)
		if (err != nil) != tt.testShouldErr {
			t.Errorf("extractRKey error for uri %q: %v, testShouldErr = %v", tt.uri, err, tt.testShouldErr)
		}

		if rkey != tt.expectedRkey {
			t.Errorf("extractRKey for uri %q = %q, expected %q", tt.uri, rkey, tt.expectedRkey)
		}

	}

}
