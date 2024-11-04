package main

import (
	"net/url"
	"reflect"
	"testing"
)

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}

func Test_findURLs(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findURLs(tt.args.s); got != tt.want {
				t.Errorf("findURLs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseURLs(t *testing.T) {
	type args struct {
		urls string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parseURLs(tt.args.urls)
		})
	}
}

func Test_removeTrackers(t *testing.T) {
	type args struct {
		u *url.URL
	}
	tests := []struct {
		name string
		args args
		want url.Values
	}{
		{"remove utm_{source,medium,name,term,content}", args{&url.URL{RawQuery: "utm_source=share&utm_medium=android_app&utm_name=androidcss&utm_term=1&utm_content=share_button"}}, url.Values{}},
		{"keep foo", args{&url.URL{RawQuery: "foo=bar"}}, url.Values{"foo": {"bar"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeTrackers(tt.args.u); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeTrackers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unfurlURL(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO commenting out for now since it api throws 403
		// {"unfurl https://t.ly", args{"https://t.ly/Ynr0"}, "https://media.dcd.uscourts.gov/datepicker/index.html"},
		{"unfurl https://ti.me", args{"https://ti.me/43d0303"}, "https://time.com/6269313/trump-jesus-comparisons-blasphemy/?utm_source=twitter&utm_medium=social&utm_campaign=editorial&utm_term=ideas_politics&linkId=208764632"},
		{"DON'T unfurl https://borretti.me", args{"https://borretti.me/about/"}, "https://borretti.me/about/"},
		// TODO the string url contains what looks like an unfurl key that might need some extra massaging
		// {"unfurl https://redd.it", args{"https://redd.it/12hs83k"}, "https://www.reddit.com/comments/12hs83k?rdt=41413"},
		{"DON'T unfurl https://i.redd.it", args{"https://i.redd.it/kzptzlwlrmu71.jpg"}, "https://i.redd.it/kzptzlwlrmu71.jpg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unfurlURL(tt.args.s); got != tt.want {
				t.Errorf("unfurlURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
