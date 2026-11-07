//go:build e2e

package tests

import "testing"

func TestEvaluationsRunConcurrently(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_closed_gate("a").and().
		an_environment().and().
		an_evaluation_is_started("slow", `t('gate', 'a')`).and().
		the_gate_is_entered("a")

	when.
		an_evaluation_runs("fast", `t('count', 'b')`)

	then.
		the_evaluation_outputs("fast", `1`).and().
		the_evaluation_is_still_running("slow")
}

func TestFlushWaitsForARunningEvaluation(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_library_file("lib.libsonnet", `{ value: 'lib' }`).and().
		a_closed_gate("a").and().
		an_environment().and().
		an_evaluation_is_started("eval", `[(import 'lib.libsonnet').value, t('gate', 'a'), (import 'lib.libsonnet').value]`).and().
		the_gate_is_entered("a")

	when.
		the_cache_is_flushed_in_the_background()

	then.
		the_flush_is_pending()
}

func TestEvaluationSpanningAFlushCompletes(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_library_file("lib.libsonnet", `{ value: 'lib' }`).and().
		a_closed_gate("a").and().
		an_environment().and().
		an_evaluation_is_started("eval", `[(import 'lib.libsonnet').value, t('gate', 'a'), (import 'lib.libsonnet').value]`).and().
		the_gate_is_entered("a").and().
		the_cache_is_flushed_in_the_background()

	when.
		the_gate_is_opened("a")

	then.
		the_evaluation_finishes("eval").and().
		the_evaluation_outputs("eval", `["lib", "a", "lib"]`).and().
		the_flush_finishes()
}

func TestPooledVMsSeeChangedFilesAfterAFlush(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_library_file("lib.libsonnet", `{ value: 'v1' }`).and().
		a_closed_gate("a").and().
		an_environment_with_pool_size(2).and().
		an_evaluation_is_started("first", `[(import 'lib.libsonnet').value, t('gate', 'a')]`).and().
		an_evaluation_is_started("second", `[(import 'lib.libsonnet').value, t('gate', 'a')]`).and().
		the_gate_is_entered("a").and().
		the_gate_is_entered("a").and().
		the_gate_is_opened("a").and().
		the_evaluation_finishes("first").and().
		the_evaluation_finishes("second").and().
		the_library_file_changes("lib.libsonnet", `{ value: 'v2' }`).and().
		the_cache_is_flushed().and().
		the_gate_is_closed("a").and().
		an_evaluation_is_started("third", `[(import 'lib.libsonnet').value, t('gate', 'a')]`).and().
		an_evaluation_is_started("fourth", `[(import 'lib.libsonnet').value, t('gate', 'a')]`).and().
		the_gate_is_entered("a").and().
		the_gate_is_entered("a")

	when.
		the_gate_is_opened("a")

	then.
		the_evaluation_finishes("third").and().
		the_evaluation_finishes("fourth").and().
		the_evaluation_outputs("third", `["v2", "a"]`).and().
		the_evaluation_outputs("fourth", `["v2", "a"]`)
}

func TestEvaluationsOnTheSameVMDoNotShareRecordedInvocations(t *testing.T) {
	given, when, then := scenario(t)

	given.
		an_open_gate("x", "x1").and().
		an_open_gate("y", "y1").and().
		a_watched_environment_with_pool_size(1).and().
		an_evaluation_runs("eval", `t('gate', 'x')`).and().
		a_watch_is_started("watch", `t('gate', 'y')`).and().
		the_watch_is_established("watch").and().
		the_gate_value_becomes("x", "x2").and().
		the_gate_value_becomes("y", "y2")

	when.
		the_invocation_changes("x")

	then.
		the_initial_watch_output_is("watch", `"y1"`).and().
		the_watch_receives_nothing("watch")
}

func TestConcurrentCacheMissesShareOneImportedFile(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_gated_library_file("lib.libsonnet", `{ value: 'lib' }`).and().
		a_closed_gate("a").and().
		an_environment().and().
		an_evaluation_is_started("first", `[(import 'lib.libsonnet').value, t('gate', 'a'), (import 'lib.libsonnet').value]`).and().
		an_evaluation_is_started("second", `[(import 'lib.libsonnet').value, t('gate', 'a'), (import 'lib.libsonnet').value]`).and().
		the_library_read_is_entered().and().
		the_library_read_is_entered().and().
		the_library_read_is_released().and().
		the_gate_is_entered("a").and().
		the_gate_is_entered("a")

	when.
		the_gate_is_opened("a")

	then.
		the_evaluation_finishes("first").and().
		the_evaluation_finishes("second").and().
		the_evaluation_outputs("first", `["lib", "a", "lib"]`).and().
		the_evaluation_outputs("second", `["lib", "a", "lib"]`)
}

func TestCancellingAnEvaluationWaitingForAFlush(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_closed_gate("a").and().
		an_environment().and().
		an_evaluation_is_started("running", `t('gate', 'a')`).and().
		the_gate_is_entered("a").and().
		the_cache_is_flushed_in_the_background().and().
		the_flush_is_pending().and().
		an_evaluation_is_started("waiting", `t('count', 'b')`)

	when.
		the_evaluation_is_cancelled("waiting")

	then.
		the_evaluation_finishes("waiting").and().
		the_evaluation_fails_with_cancellation("waiting").and().
		the_count_is("b", 0).and().
		the_flush_is_pending()
}

func TestAlreadyCancelledEvaluationDoesNotWaitForAFlush(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_closed_gate("a").and().
		an_environment().and().
		an_evaluation_is_started("running", `t('gate', 'a')`).and().
		the_gate_is_entered("a").and().
		the_cache_is_flushed_in_the_background().and().
		the_flush_is_pending()

	when.
		an_already_cancelled_evaluation_is_started("cancelled", `t('count', 'b')`)

	then.
		the_evaluation_finishes("cancelled").and().
		the_evaluation_fails_with_cancellation("cancelled").and().
		the_count_is("b", 0)
}
