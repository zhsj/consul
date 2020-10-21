package inspect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

const (
	PrettyFormat string = "pretty"
	JSONFormat   string = "json"
)

type Formatter interface {
	Format(*OutputFormat) (string, error)
}

func GetSupportedFormats() []string {
	return []string{PrettyFormat, JSONFormat}
}

type prettyFormatter struct{}

func newPrettyFormatter() Formatter {
	return &prettyFormatter{}
}
func NewFormatter(format string) (Formatter, error) {
	switch format {
	case PrettyFormat:
		return newPrettyFormatter(), nil
	case JSONFormat:
		return newJSONFormatter(), nil
	default:
		return nil, fmt.Errorf("Unknown format: %s", format)
	}
}

func (_ *prettyFormatter) Format(info *OutputFormat) (string, error) {
	var b bytes.Buffer
	// For the enhanced stats
	ss := make([]typeStats, 0, len(info.Stats))

	for _, s := range info.Stats {
		ss = append(ss, s)
	}

	// Sort the stat slice
	sort.Slice(ss, func(i, j int) bool { return ss[i].Sum > ss[j].Sum })
	tw := tabwriter.NewWriter(&b, 8, 8, 6, ' ', 0)

	fmt.Fprintf(tw, " ID\t%s", info.Meta.ID)
	fmt.Fprintf(tw, "\n Size\t%d", info.Meta.Size)
	fmt.Fprintf(tw, "\n Index\t%d", info.Meta.Index)
	fmt.Fprintf(tw, "\n Term\t%d", info.Meta.Term)
	fmt.Fprintf(tw, "\n Version\t%d", info.Meta.Version)
	fmt.Fprintf(tw, "\n")
	fmt.Fprintln(tw, "\n Type\tCount\tSize\t")
	fmt.Fprintf(tw, " %s\t%s\t%s\t", "----", "----", "----")
	// For each different type generate new output
	for _, s := range ss {
		fmt.Fprintf(tw, "\n %s\t%d\t%s\t", s.Name, s.Count, ByteSize(uint64(s.Sum)))
	}
	fmt.Fprintf(tw, "\n %s\t%s\t%s\t", "----", "----", "----")
	fmt.Fprintf(tw, "\n Total\t\t%s\t", ByteSize(uint64(info.Offset)))

	if err := tw.Flush(); err != nil {
		return b.String(), err
	}
	return b.String(), nil
}

type jsonFormatter struct{}

func newJSONFormatter() Formatter {
	return &jsonFormatter{}
}

func (_ *jsonFormatter) Format(info *OutputFormat) (string, error) {
	b, err := json.MarshalIndent(info, "", "   ")
	if err != nil {
		return "", fmt.Errorf("Failed to marshal original snapshot stats: %v", err)
	}
	return string(b), nil
}

const (
	BYTE = 1 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	TERABYTE
)

func ByteSize(bytes uint64) string {
	unit := ""
	value := float64(bytes)

	switch {
	case bytes >= TERABYTE:
		unit = "TB"
		value = value / TERABYTE
	case bytes >= GIGABYTE:
		unit = "GB"
		value = value / GIGABYTE
	case bytes >= MEGABYTE:
		unit = "MB"
		value = value / MEGABYTE
	case bytes >= KILOBYTE:
		unit = "KB"
		value = value / KILOBYTE
	case bytes >= BYTE:
		unit = "B"
	case bytes == 0:
		return "0"
	}

	result := strconv.FormatFloat(value, 'f', 1, 64)
	result = strings.TrimSuffix(result, ".0")
	return result + unit
}