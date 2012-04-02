// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ics provides support for reading Apple's iCalendar file format.
package ics

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"sort"
	"strings"
	"time"
)

type Calendar struct {
	Event []*Event
}

type Event struct {
	UID                            string
	Start, End                     time.Time
	Summary, Location, Description string
}

func Decode(rd io.Reader) (c *Calendar, err error) {
	r := bufio.NewReader(rd)
	for {
		key, value, err := decodeLine(r)
		if err != nil {
			return nil, err
		}
		if key == "BEGIN" {
			if c == nil {
				if value != "VCALENDAR" {
					return nil, errors.New("didn't find BEGIN:VCALENDAR")
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

func decodeEvent(r *bufio.Reader) (*Event, error) {
	e := new(Event)
	var key, value string
	var err error
	for {
		if err != nil {
			return nil, err
		}
		key, value, err = decodeLine(r)
		switch key {
		case "END":
			if value != "VEVENT" {
				return nil, errors.New("unexpected END value")
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

func decodeTime(value string) (time.Time, error) {
	const layout = "20060102T150405Z"
	return time.Parse(layout, value)
}

func decodeLine(r *bufio.Reader) (key, value string, err error) {
	var buf bytes.Buffer
	for {
		// get full line
		b, isPrefix, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return "", "", err
		}
		if isPrefix {
			return "", "", errors.New("unexpected long line")
		}
		if len(b) == 0 {
			return "", "", errors.New("unexpected blank line")
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
		return "", "", errors.New("bad line, couldn't find key:value")
	}
	return p[0], p[1], nil
}

type eventList []*Event

func (l eventList) Less(i, j int) bool {
	if l[i].Start.IsZero() {
		return true
	}
	if l[j].Start.IsZero() {
		return false
	}
	return l[i].Start.Before(l[j].Start)
}
func (l eventList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l eventList) Len() int      { return len(l) }
