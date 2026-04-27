package cppbridge

import (
	"fmt"
	"os"
	"path/filepath"
)

// BridgeConfig 定义 C++ 推荐桥的初始化参数。
type BridgeConfig struct {
	ConfigPath       string
	UniversitiesPath string
	MajorsPath       string
	HistoricalPath   string
}

type bridgeDataFiles struct {
	universities string
	majors       string
	historical   string
}

func resolveBridgeDataFiles(cfg BridgeConfig) (bridgeDataFiles, error) {
	universities, err := resolveOptionalPath(cfg.UniversitiesPath,
		"/app/data/universities.csv",
		"./data/universities.csv",
		"cpp-modules/volunteer-matcher/data/universities.csv",
		"../../cpp-modules/volunteer-matcher/data/universities.csv",
		"../../../../cpp-modules/volunteer-matcher/data/universities.csv",
	)
	if err != nil {
		return bridgeDataFiles{}, fmt.Errorf("resolve universities data file: %w", err)
	}

	majors, err := resolveOptionalPath(cfg.MajorsPath,
		"/app/data/majors.csv",
		"./data/majors.csv",
		"cpp-modules/volunteer-matcher/data/majors.csv",
		"../../cpp-modules/volunteer-matcher/data/majors.csv",
		"../../../../cpp-modules/volunteer-matcher/data/majors.csv",
	)
	if err != nil {
		return bridgeDataFiles{}, fmt.Errorf("resolve majors data file: %w", err)
	}

	historical, err := resolveOptionalPath(cfg.HistoricalPath,
		"/app/data/historical_data.csv",
		"./data/historical_data.csv",
		"cpp-modules/volunteer-matcher/data/historical_data.csv",
		"../../cpp-modules/volunteer-matcher/data/historical_data.csv",
		"../../../../cpp-modules/volunteer-matcher/data/historical_data.csv",
	)
	if err != nil {
		return bridgeDataFiles{}, fmt.Errorf("resolve historical data file: %w", err)
	}

	return bridgeDataFiles{
		universities: universities,
		majors:       majors,
		historical:   historical,
	}, nil
}

func resolveOptionalPath(preferred string, candidates ...string) (string, error) {
	paths := candidates
	if preferred != "" {
		paths = append([]string{preferred}, candidates...)
	}

	for _, candidate := range paths {
		if candidate == "" {
			continue
		}
		cleaned := filepath.Clean(candidate)
		if _, err := os.Stat(cleaned); err == nil {
			return cleaned, nil
		}
	}

	return "", fmt.Errorf("no existing file found in candidates: %v", paths)
}
