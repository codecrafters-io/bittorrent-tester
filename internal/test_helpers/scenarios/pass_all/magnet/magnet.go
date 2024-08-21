package magnet

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

// Magnet is a parsed URI.
type MagnetLink struct {
	InfoHash string
	// DisplayName is a "dp".
	DisplayName string
	// ExactTopics is a "xt".
	ExactTopics []string
	// ExactLength is a "xl".
	ExactLength int64
	// AcceptableSources is "as".
	AcceptableSources []string
	// ExactSource is a "xs".
	ExactSource []string
	// KeywordTopic is a "kt".
	KeywordTopic []string
	// ManifestTopic is a "mt".
	ManifestTopic string
	// Trackers is a "tr".
	Trackers []string
	// AdditionParams is a "x." and other unparsed params.
	AdditionParams map[string][]string
}

var ErrUnsupportedFormat = errors.New("uri doesn't starts from prefix")

func Parse(uri string) (*MagnetLink, error) {
	if !strings.HasPrefix(uri, "magnet:?") {
		return nil, ErrUnsupportedFormat
	}
	// skip magnet:?
	parts := strings.Split(uri[8:], "&")
	if len(parts) < 1 {
		return nil, ErrUnsupportedFormat
	}
	magnet := parts[0]
	hash := strings.Split(magnet, ":")
	trackers := []string{}
	m := MagnetLink{
		AdditionParams: make(map[string][]string),
	}
	if len(hash) > 2 {
		m.InfoHash = hash[2]
	}
	exactTopics := make(map[string]struct{})
	acceptableSources := make(map[string]struct{})

	for _, part := range parts {
		partSplit := strings.Split(part, "=")
		if len(partSplit) < 2 || partSplit[1] == "" {
			continue
		}
		key := partSplit[0]
		value := partSplit[1]
		switch key {
		case "tr":
			decoded, err := url.QueryUnescape(value)
			if err != nil {
				return nil, err
			}
			trackers = append(trackers, decoded)
		case "dn":
			decoded, err := url.QueryUnescape(value)
			if err != nil {
				return nil, err
			}
			m.DisplayName = decoded
		case "xt":
			if _, exist := exactTopics[value]; !exist {
				exactTopics[value] = struct{}{}
			}
		case "mt":
			m.ManifestTopic = value
		case "xl":
			size, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, err
			}
			m.ExactLength = size
		case "as":
			decoded, err := url.QueryUnescape(value)
			if err != nil {
				return nil, err
			}

			if _, exist := acceptableSources[decoded]; !exist {
				acceptableSources[decoded] = struct{}{}
			}
		case "xs":
			decoded, err := url.QueryUnescape(value)
			if err != nil {
				return nil, err
			}

			m.ExactSource = append(m.ExactSource, decoded)
		case "kt":
			m.KeywordTopic = append(m.KeywordTopic, strings.Split(value, "+")...)
		default:
			decoded, err := url.QueryUnescape(value)
			if err != nil {
				return nil, err
			}

			m.AdditionParams[key] = append(m.AdditionParams[key], decoded)
		}

	}
	m.ExactTopics = make([]string, 0, len(exactTopics))
	for topic := range exactTopics {
		m.ExactTopics = append(m.ExactTopics, topic)
	}
	m.AcceptableSources = make([]string, 0, len(acceptableSources))
	for as := range acceptableSources {
		m.AcceptableSources = append(m.AcceptableSources, as)
	}
	m.Trackers = trackers
	return &m, nil
}

func (link *MagnetLink) Encode() string {
	return ""
}
