package telemetry

import "github.com/samuelnp/centinela/internal/config"

// RecordBlock records a prewrite block. feature/step are empty on a need-init
// block; reason is "need-init" or "out-of-step".
func RecordBlock(cfg *config.Config, feature, step, fileType, targetPath, reason string) {
	Record(cfg, Event{
		Type:       TypeBlock,
		Feature:    feature,
		Step:       step,
		Reason:     reason,
		FileType:   fileType,
		TargetPath: targetPath,
	})
}

// RecordGateFailure records one failing validate gate.
func RecordGateFailure(cfg *config.Config, gate, message string) {
	Record(cfg, Event{Type: TypeGateFailure, Gate: gate, Message: message})
}

// RecordVerifyRejection records a hard-blocking claim verification failure.
func RecordVerifyRejection(cfg *config.Config, feature, step string, checks []CheckRef) {
	Record(cfg, Event{Type: TypeVerifyRejection, Feature: feature, Step: step, Checks: checks})
}

// RecordCompleteRejected records a refused advance; reason is "gates" or "verify".
func RecordCompleteRejected(cfg *config.Config, feature, step, reason string) {
	Record(cfg, Event{Type: TypeCompleteRejected, Feature: feature, Step: step, Reason: reason})
}

// RecordStepAdvanced records a successful advance (brackets the rework window),
// carrying the just-completed step.
func RecordStepAdvanced(cfg *config.Config, feature, step string) {
	Record(cfg, Event{Type: TypeStepAdvanced, Feature: feature, Step: step})
}
