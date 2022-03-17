package logger

import "go.uber.org/zap"

type noop struct{}

func (n *noop) Debug(msg string, fields ...zap.Field) {}

func (n *noop) Info(msg string, fields ...zap.Field) {}

func (n *noop) Warn(msg string, fields ...zap.Field) {}

func (n *noop) Error(msg string, fields ...zap.Field) {}
