package internal

import (
	logger "github.com/codecrafters-io/tester-utils/logger"
)

func logOnExit(logger *logger.Logger, err *error) {
	if *err != nil {
		logger.Errorf("%v", *err)
	}
}
