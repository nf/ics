// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ics provides support for reading Apple's iCalendar file format.
package ics

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"time"
	"sort"
	"strings"
)

type Calendar struct {
	Event []*Event
}

type Event struct {
	UID                            string
	Start, End                     *time.Time
	Summary, Location, Description string
}

func Decode(rd io.Reader) (c *Calendar, err os.Error) {
	r := bufio.NewReader(rd)
	for {
		key, value, err := decodeLine(r)
		if err != nil {
			return nil, err
		}
		if key == "BEGIN" {
			if c == nil {
				if value != "VCALENDAR" {
					return nil, os.NewError("didn't find BEGIN:VCALENDAR")
				}
				c = new(Calendar)
			}
			if value == "VEVENT" {
				e, err := decodeEvent(r)
				if err != nil {
					return nil, err
				}
				c.Event = append(c.Event, e)
			}
		}
		if key == "END" && value == "VCALENDAR" {
			break
		}
	}
	sort.Sort(eventList(c.Event))
	return c, nil
}

func decodeEvent(r *bufio.Reader) (*Event, os.Error) {
	e := new(Event)
	var key, value string
	var err os.Error
	for {
		if err != nil {
			return nil, err
		}
		key, value, err = decodeLine(r)
		switch key {
		case "END":
			if value != "VEVENT" {
				return nil, os.NewError("unexpected END value")
			}
			return e, nil
		case "UID":
			e.UID = value
		case "DTSTART":
			e.Start, err = decodeTime(value)
		case "DTEND":
			e.End, err = decodeTime(value)
		case "SUMMARY":
			e.Summary = value
		case "LOCATION":
			e.Location = value
		case "DESCRIPTION":
			e.Description = value
		}
	}
	panic("unreachable")
}

func decodeTime(value string) (*time.Time, os.Error) {
	const layout = "20060102T150405Z"
	return time.Parse(layout, value)
}

func decodeLine(r *bufio.Reader) (key, value string, err os.Error) {
	var buf bytes.Buffer
	for {
		// get full line
		b, isPrefix, err := r.ReadLine()
		if err != nil {
			if err == os.EOF {
				err = io.ErrUnexpectedEOF
			}
			return "", "", err
		}
		if isPrefix {
			return "", "", os.NewError("unexpected long line")
		}
		if len(b) == 0 {
			return "", "", os.NewError("unexpected blank line")
		}
		if b[0] == ' ' {
			b = b[1:]
		}
		buf.Write(b)

		b, err = r.Peek(1)
		if err != nil || b[0] != ' ' {
			break
		}
	}
	p := strings.SplitN(buf.String(), ":", 2)
	if len(p) != 2 {
		return "", "", os.NewError("bad line, couldn't find key:value")
	}
	return p[0], p[1], nil
}

type eventList []*Event

func (l eventList) Less(i, j int) bool {
	if l[i].Start == nil || l[j].Start == nil {
		return true
	}
	return l[i].Start.Seconds() < l[j].Start.Seconds()
}
func (l eventList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l eventList) Len() int      { return len(l) }
