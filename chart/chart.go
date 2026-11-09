package chart

import (
	"bufio"
	"fmt"
	"nsgo/msg"
	"nsgo/tools"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/go-analyze/charts"
)

type PacketRecord struct {
	Time      float64
	SizeBytes float64
	SrcIP     string
}

func Graph(filename string, dest string) tools.Error {
	records, err := readTR(filename)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return tools.NewErrorf(tools.ErrChartGeneral, "no server-received records found in %s", filename)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Time < records[j].Time
	})

	const binSize = 1.0
	maxTime := records[len(records)-1].Time

	numBins := int(maxTime/binSize) + 1
	if numBins < 1 {
		numBins = 1
	}

	binLabels := make([]string, numBins)
	bytesPerBin := make([]float64, numBins)
	packetsPerBin := make([]float64, numBins)

	for i := 0; i < numBins; i++ {
		binLabels[i] = fmt.Sprintf("%.0f", float64(i)*binSize)
	}

	for _, rec := range records {
		binIdx := int(rec.Time / binSize)
		if binIdx >= numBins {
			binIdx = numBins - 1
		}
		bytesPerBin[binIdx] += rec.SizeBytes
		packetsPerBin[binIdx]++
	}

	cumulativeBytes := make([]float64, numBins)
	var runningTotal float64
	for i := 0; i < numBins; i++ {
		runningTotal += bytesPerBin[i]
		cumulativeBytes[i] = runningTotal
	}

	throughputKBps := make([]float64, numBins)
	for i := 0; i < numBins; i++ {
		throughputKBps[i] = bytesPerBin[i] / 1024.0
	}

	packetRate := make([]float64, numBins)
	copy(packetRate, packetsPerBin)

	if err := renderChart(
		dest+"/cumulative_received.png",
		"Cumulative Received Data",
		"Time (s)",
		"Bytes",
		[][]float64{cumulativeBytes},
		[]string{"Cumulative Bytes"},
		binLabels,
	); err != nil {
		return err
	}

	if err := renderChart(
		dest+"/throughput.png",
		"Throughput Over Time",
		"Time (s)",
		"KB/s",
		[][]float64{throughputKBps},
		[]string{"Throughput (KB/s)"},
		binLabels,
	); err != nil {
		return err
	}

	if err := renderChart(
		dest+"/packet_rate.png",
		"Packet Rate Over Time",
		"Time (s)",
		"Packets/s",
		[][]float64{packetRate},
		[]string{"Packet Rate (pkt/s)"},
		binLabels,
	); err != nil {
		return err
	}

	msg.SUCCESSln("Charts generated: cumulative_received.png, throughput.png, packet_rate.png")
	return nil
}

func readTR(path string) ([]PacketRecord, tools.Error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, tools.NewErrorf(tools.ErrChartGeneral, "opening %s: %w", path, err)
	}
	defer f.Close()

	var records []PacketRecord
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] != 'r' {
			continue
		}
		rec, ok := parseRecvLine(line)
		if !ok {
			continue
		}
		if !strings.HasSuffix(rec.SrcIP, ".2") {
			continue
		}
		records = append(records, rec)
	}

	if err := scanner.Err(); err != nil {
		return nil, tools.NewErrorf(tools.ErrChartGeneral, "scanning %s: %w", path, err)
	}
	return records, nil
}

func parseRecvLine(line string) (PacketRecord, bool) {
	var rec PacketRecord
	tokens := strings.Fields(line)
	if len(tokens) < 2 || tokens[0] != "r" {
		return rec, false
	}

	t, err := strconv.ParseFloat(tokens[1], 64)
	if err != nil {
		return rec, false
	}
	rec.Time = t

	var srcIP string
	var sizeBytes float64
	var sizeFound bool

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		if tok == ">" && i > 0 && srcIP == "" {
			if looksLikeIPv4(tokens[i-1]) {
				srcIP = tokens[i-1]
			}
			continue
		}

		if !sizeFound && strings.HasPrefix(tok, "size=") {
			s := strings.TrimRight(strings.TrimPrefix(tok, "size="), ")")
			if v, err := strconv.ParseFloat(s, 64); err == nil {
				sizeBytes = v
				sizeFound = true
			}
			continue
		}

		if !sizeFound && tok == "(size=" && i+1 < len(tokens) {
			s := strings.TrimRight(tokens[i+1], ")")
			if v, err := strconv.ParseFloat(s, 64); err == nil {
				sizeBytes = v
				sizeFound = true
			}
			continue
		}

		if !sizeFound && tok == "length:" && i+1 < len(tokens) {
			if v, err := strconv.ParseFloat(tokens[i+1], 64); err == nil {
				sizeBytes = v
				sizeFound = true
			}
			continue
		}
	}

	if !sizeFound || srcIP == "" {
		return rec, false
	}
	rec.SizeBytes = sizeBytes
	rec.SrcIP = srcIP
	return rec, true
}

func looksLikeIPv4(s string) bool {
	if len(s) < 7 || len(s) > 15 {
		return false
	}
	parts := 0
	digits := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			digits++
			if digits > 3 {
				return false
			}
		case c == '.':
			if digits == 0 {
				return false
			}
			parts++
			digits = 0
		default:
			return false
		}
	}
	return parts == 3 && digits > 0
}

func renderChart(filename, title, xLabel, yLabel string, values [][]float64, seriesNames []string, xLabels []string) tools.Error {
	opt := charts.NewLineChartOptionWithData(values)
	opt.Title = charts.TitleOption{Text: title}
	opt.XAxis = charts.XAxisOption{Labels: xLabels}
	opt.YAxis = []charts.YAxisOption{{Title: yLabel}}
	opt.Legend = charts.LegendOption{SeriesNames: seriesNames}

	p := charts.NewPainter(charts.PainterOptions{
		Width:  1200,
		Height: 500,
	})

	if err := p.LineChart(opt); err != nil {
		return tools.NewErrorf(tools.ErrChartGeneral, "rendering %s: %w", filename, err)
	}

	buf, err := p.Bytes()
	if err != nil {
		return tools.NewErrorf(tools.ErrChartGeneral, "rendering %s: %w", filename, err)
	}

	if err := os.WriteFile(filename, buf, 0644); err != nil {
		return tools.NewErrorf(tools.ErrChartGeneral, "writing %s: %w", filename, err)
	}
	return nil
}
