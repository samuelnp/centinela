package telemetry

import "github.com/samuelnp/centinela/internal/config"

// RecordBlock records a prewrite block. feature/step are empty on a need-init
// block; reason is "need-init" or "out-of-step". model is the resolved driver
// model id (resolved by the cmd/ caller so telemetry stays a config-only leaf).
func RecordBlock(cfg *config.Config, feature, step, fileType, targetPath, reason, model string) {
	Record(cfg, Event{
		Type:       TypeBlock,
		Feature:    feature,
		Step:       step,
		Model:      model,
		Reason:     reason,
		FileType:   fileType,
		TargetPath: targetPath,
	})
}

// RecordGateFailure records one failing validate gate.
func RecordGateFailure(cfg *config.Config, gate, message, model string) {
	Record(cfg, Event{Type: TypeGateFailure, Gate: gate, Message: message, Model: model})
}

// RecordVerifyRejection records a hard-blocking claim verification failure.
func RecordVerifyRejection(cfg *config.Config, feature, step string, checks []CheckRef, model string) {
	Record(cfg, Event{Type: TypeVerifyRejection, Feature: feature, Step: step, Model: model, Checks: checks})
}

// RecordCompleteRejected records a refused advance; reason is "gates" or "verify".
func RecordCompleteRejected(cfg *config.Config, feature, step, reason, model string) {
	Record(cfg, Event{Type: TypeCompleteRejected, Feature: feature, Step: step, Model: model, Reason: reason})
}

// RecordStepAdvanced records a successful advance (brackets the rework window),
// carrying the just-completed step.
func RecordStepAdvanced(cfg *config.Config, feature, step, model string) {
	Record(cfg, Event{Type: TypeStepAdvanced, Feature: feature, Step: step, Model: model})
}

// RecordCostSample records host-harness token spend attributed to the active
// feature/step/model. No-op for a non-positive total so an empty transcript
// delta never writes a noise line.
func RecordCostSample(cfg *config.Config, feature, step, model string, in, out int) {
	if in <= 0 && out <= 0 {
		return
	}
	Record(cfg, Event{Type: TypeCostSample, Feature: feature, Step: step, Model: model, InputTokens: in, OutputTokens: out})
}
