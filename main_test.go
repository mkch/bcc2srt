package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestParseBcc(t *testing.T) {
	data := `{
    "font_size": 0.4,
    "font_color": "#FFFFFF",
    "background_alpha": 0.5,
    "background_color": "#9C27B0",
    "Stroke": "none",
    "body": [
        {
            "from": 0.3,
            "to": 0.66,
            "location": 2,
            "content": "大家好"
        },
        {
            "from": 0.66,
            "to": 2.25,
            "location": 2,
            "content": "我是 自觉放齐"
        }
	]
}`

	if ret, err := parseBcc(strings.NewReader(data)); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(ret, &bcc{
		Body: []struct {
			From    float32 `json:"from"`
			To      float32 `json:"to"`
			Content string  `json:"content"`
		}{{0.3, 0.66, "大家好"}, {0.66, 2.25, "我是 自觉放齐"}},
	}) {
		t.Fatal(ret)
	}
}

func TestSrtTime(t *testing.T) {
	if str := srtTime(0.3); str != "00:00:00,300" {
		t.Fatal(str)
	}
	if str := srtTime(171.92); str != "00:02:51,920" {
		t.Fatal(str)
	}
}

func TestBcc2Srt(t *testing.T) {
	const bcc = `{
    "font_size": 0.4,
    "font_color": "#FFFFFF",
    "background_alpha": 0.5,
    "background_color": "#9C27B0",
    "Stroke": "none",
    "body": [
        {
            "from": 0.3,
            "to": 0.66,
            "location": 2,
            "content": "大家好"
        },
        {
            "from": 0.66,
            "to": 2.25,
            "location": 2,
            "content": "我是 自觉放齐"
        }
	]
}`

	const srt = `1
00:00:00,300 --> 00:00:00,660
大家好

2
00:00:00,660 --> 00:00:02,250
我是 自觉放齐

`

	var w bytes.Buffer
	err := bcc2srt(strings.NewReader(bcc), &w)
	if err != nil {
		t.Fatal(err)
	}

	if ret := w.String(); ret != srt {
		t.Fatal(ret)
	}
}

func Test_changeExt(t *testing.T) {
	type args struct {
		filename string
		ext      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal", args{"a.txt", "html"}, "a.html"},
		{"no_ext", args{"a", "html"}, "a.html"},
		{"no_filename", args{".txt", "html"}, ".html"},
		{"empty", args{"", "html"}, ".html"},
		{"empty_ext", args{"a.txt", ""}, "a."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := changeExt(tt.args.filename, tt.args.ext); got != tt.want {
				t.Errorf("changeExt() = %v, want %v", got, tt.want)
			}
		})
	}
}
