package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ResolveOutputName(inputPath string, opSuffix string, targetExt string, explicitOutput string, allowOverwrite bool) (string, error) {
	if explicitOutput != "" {
		if !allowOverwrite {
			if _, err := os.Stat(explicitOutput); err == nil {
				return "", fmt.Errorf("output file %q already exists; use -y to overwrite", explicitOutput)
			}
		}
		return explicitOutput, nil
	}

	dir := filepath.Dir(inputPath)
	base := filepath.Base(inputPath)
	origExt := filepath.Ext(base)
	stem := strings.TrimSuffix(base, origExt)

	targetExt = strings.TrimPrefix(targetExt, ".")
	if targetExt == "" {
		targetExt = strings.TrimPrefix(origExt, ".")
	}

	var baseName string
	if opSuffix != "" {
		baseName = fmt.Sprintf("%s.%s.%s", stem, opSuffix, targetExt)
	} else {
		baseName = fmt.Sprintf("%s.%s", stem, targetExt)
	}

	candidate := filepath.Join(dir, baseName)
	if allowOverwrite {
		return candidate, nil
	}

	if _, err := os.Stat(candidate); os.IsNotExist(err) {
		return candidate, nil
	}

	for i := 1; i <= 999; i++ {
		var altName string
		if opSuffix != "" {
			altName = fmt.Sprintf("%s.%s.%d.%s", stem, opSuffix, i, targetExt)
		} else {
			altName = fmt.Sprintf("%s.%d.%s", stem, i, targetExt)
		}
		altCandidate := filepath.Join(dir, altName)
		if _, err := os.Stat(altCandidate); os.IsNotExist(err) {
			return altCandidate, nil
		}
	}

	return "", fmt.Errorf("unable to find available output filename for %q", inputPath)
}
