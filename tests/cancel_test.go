//go:build e2e

package tests

import "testing"

func TestCancelledEvaluationSkipsFurtherPluginCalls(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_closed_gate("a").and().
		an_environment().and().
		an_evaluation_is_started("eval", `[t('hold', 'a'), t('count', 'b')]`).and().
		the_gate_is_entered("a").and().
		the_evaluation_is_cancelled("eval")

	when.
		the_gate_is_opened("a")

	then.
		the_evaluation_finishes("eval").and().
		the_evaluation_fails_with_cancellation("eval").and().
		the_count_is("b", 0)
}

func TestCancellingReachesARunningInProcessPluginCall(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_closed_gate("a").and().
		an_environment().and().
		an_evaluation_is_started("eval", `t('gate', 'a')`).and().
		the_gate_is_entered("a")

	when.
		the_evaluation_is_cancelled("eval")

	then.
		the_gate_observes_cancellation("a").and().
		the_evaluation_finishes("eval").and().
		the_evaluation_fails_with_cancellation("eval")
}

func TestCancellingReachesARunningOutOfProcessPluginCall(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_fixture_plugin().and().
		an_environment().and().
		a_fixture_gate_evaluation_is_started("eval").and().
		the_fixture_gate_is_entered()

	when.
		the_evaluation_is_cancelled("eval")

	then.
		the_fixture_gate_observes_cancellation().and().
		the_evaluation_finishes("eval").and().
		the_evaluation_fails_with_cancellation("eval")
}

func TestCancelledEvaluationFailsEvenWhenThePluginReturnsAFailure(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_closed_gate("a").and().
		an_environment().and().
		an_evaluation_is_started("eval", `t('soft', 'a')`).and().
		the_gate_is_entered("a")

	when.
		the_evaluation_is_cancelled("eval")

	then.
		the_gate_observes_cancellation("a").and().
		the_evaluation_finishes("eval").and().
		the_evaluation_fails_with_cancellation("eval")
}

func TestCancellingAnEvaluationWaitingForAVM(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_closed_gate("a").and().
		an_environment_with_pool_size(1).and().
		an_evaluation_is_started("running", `t('gate', 'a')`).and().
		the_gate_is_entered("a").and().
		an_evaluation_is_started("waiting", `t('count', 'b')`)

	when.
		the_evaluation_is_cancelled("waiting")

	then.
		the_evaluation_finishes("waiting").and().
		the_evaluation_fails_with_cancellation("waiting").and().
		the_count_is("b", 0).and().
		the_evaluation_is_still_running("running")
}
