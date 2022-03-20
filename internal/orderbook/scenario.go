package orderbook

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Scenario struct {
	Description  string
	ShouldTrade  bool
	Instructions string
	Output       string
}

// GetScenarios gets all scenarios from input.txt and associated output (output.txt)
func GetScenarios(folderPath string) ([]*Scenario, error) {
	var scenarios []*Scenario

	// Input
	b, err := os.ReadFile(fmt.Sprintf("%s/input.txt", folderPath))
	if err != nil {
		return nil, err
	}

	line := 0
	for _, s := range bytes.Split(bytes.TrimPrefix(b, []byte(`# `)), []byte("\n# ")) {
		if line == 0 {
			line++
			continue
		}

		header, inst, _ := bytes.Cut(s, []byte{'\n'})
		shouldTradeBytes, desc, _ := bytes.Cut(header, []byte{' '})
		shouldTrade := string(shouldTradeBytes) == "1"
		x := &Scenario{
			Description:  string(desc),
			ShouldTrade:  shouldTrade,
			Instructions: string(inst),
		}
		scenarios = append(scenarios, x)
	}

	// Output
	b, err = os.ReadFile(fmt.Sprintf("%s/output.txt", folderPath))
	if err != nil {
		return nil, err
	}

	var outputs []string
	for _, s := range bytes.Split(bytes.TrimPrefix(b, []byte(`# `)), []byte("\n# ")) {
		_, output, _ := bytes.Cut(s, []byte{'\n'})
		outputs = append(outputs, strings.TrimSpace(string(output)))
	}

	if len(outputs) != len(scenarios) {
		return nil, errors.New("Inputs and outputs number should be the same")
	}

	for i, o := range outputs {
		// Remove '\n'
		if o[len(o)-1] == '\n' {
			o = o[:len(o)-1]
		}
		scenarios[i].Output = o
	}

	return scenarios, nil
}
