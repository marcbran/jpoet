//go:build e2e

package tests

import "testing"

func TestConcurrentWatchesTrackTheirOwnInvocations(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_closed_gate("a").and().
		an_open_gate("b", "b1").and().
		a_watched_environment().and().
		a_watch_is_started("watch-a", `t('gate', 'a')`).and().
		the_gate_is_entered("a").and().
		a_watch_is_started("watch-b", `t('gate', 'b')`).and().
		the_watch_is_established("watch-b").and().
		the_gate_is_opened("a").and().
		the_watch_is_established("watch-a").and().
		the_gate_value_becomes("b", "b2")

	when.
		the_invocation_changes("b")

	then.
		the_initial_watch_output_is("watch-a", `"a"`).and().
		the_initial_watch_output_is("watch-b", `"b1"`).and().
		the_watch_receives("watch-b", `"b2"`).and().
		the_watch_receives_nothing("watch-a")
}

func TestSupersededWatchEvaluationIsCancelled(t *testing.T) {
	given, when, then := scenario(t)

	given.
		an_open_gate("a", "v1").and().
		a_watched_environment().and().
		a_watch_is_started("watch", `t('gate', 'a')`).and().
		the_watch_is_established("watch").and().
		the_gate_is_closed("a").and().
		the_invocation_changes("a").and().
		the_gate_is_entered("a")

	when.
		the_invocation_changes("a")

	then.
		the_gate_observes_cancellation("a")
}

func TestOnlyTheLatestWatchEvaluationIsDelivered(t *testing.T) {
	given, when, then := scenario(t)

	given.
		an_open_gate("a", "v1").and().
		a_watched_environment().and().
		a_watch_is_started("watch", `t('gate', 'a')`).and().
		the_watch_is_established("watch").and().
		the_gate_is_closed("a").and().
		the_invocation_changes("a").and().
		the_gate_is_entered("a").and().
		the_gate_value_becomes("a", "v2").and().
		the_invocation_changes("a").and().
		the_gate_observes_cancellation("a").and().
		the_gate_is_entered("a")

	when.
		the_gate_is_opened("a")

	then.
		the_initial_watch_output_is("watch", `"v1"`).and().
		the_watch_receives("watch", `"v2"`).and().
		the_watch_receives_nothing("watch")
}

func TestUnsubscribingCancelsARunningWatchEvaluation(t *testing.T) {
	given, when, then := scenario(t)

	given.
		an_open_gate("a", "v1").and().
		a_watched_environment().and().
		a_watch_is_started("watch", `t('gate', 'a')`).and().
		the_watch_is_established("watch").and().
		the_gate_is_closed("a").and().
		the_invocation_changes("a").and().
		the_gate_is_entered("a")

	when.
		the_watch_is_unsubscribed("watch")

	then.
		the_gate_observes_cancellation("a").and().
		the_watch_is_closed_without_updates("watch")
}

func TestFileChangeCancelsARunningWatchEvaluation(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_library_file("lib.libsonnet", `{ value: 'v1' }`).and().
		an_open_gate("a", "a").and().
		a_watched_environment().and().
		a_watch_is_started("watch", `[(import 'lib.libsonnet').value, t('gate', 'a')]`).and().
		the_watch_is_established("watch").and().
		the_gate_is_closed("a").and().
		the_invocation_changes("a").and().
		the_gate_is_entered("a")

	when.
		the_library_file_changes("lib.libsonnet", `{ value: 'v2' }`)

	then.
		the_gate_observes_cancellation("a")
}

func TestClosingTheWatchedEnvironmentCancelsARunningWatchEvaluation(t *testing.T) {
	given, when, then := scenario(t)

	given.
		an_open_gate("a", "v1").and().
		a_watched_environment().and().
		a_watch_is_started("watch", `t('gate', 'a')`).and().
		the_watch_is_established("watch").and().
		the_gate_is_closed("a").and().
		the_invocation_changes("a").and().
		the_gate_is_entered("a")

	when.
		the_watched_environment_is_closed_in_the_background()

	then.
		the_gate_observes_cancellation("a").and().
		the_close_finishes()
}

func TestCancelledSubscriberDoesNotSwallowAnUpdateForExistingSubscribers(t *testing.T) {
	given, when, then := scenario(t)

	given.
		an_open_gate("a", "v1").and().
		a_watched_environment().and().
		a_watch_is_started("watch", `t('gate', 'a')`).and().
		the_watch_is_established("watch").and().
		the_gate_is_closed("a").and().
		the_invocation_changes("a").and().
		the_gate_is_entered("a").and().
		the_gate_value_becomes("a", "v2").and().
		an_already_cancelled_subscriber_joins("late", "watch", `t('gate', 'a')`).and().
		the_subscription_fails_with_cancellation("late").and().
		the_watch_is_unsubscribed("late")

	when.
		the_gate_is_opened("a")

	then.
		the_watch_receives("watch", `"v2"`)
}

func TestJoiningSubscriberDoesNotSwallowAnUpdateForExistingSubscribers(t *testing.T) {
	given, when, then := scenario(t)

	given.
		an_open_gate("a", "v1").and().
		a_watched_environment().and().
		a_watch_is_started("watch", `t('gate', 'a')`).and().
		the_watch_is_established("watch").and().
		the_gate_is_closed("a").and().
		the_invocation_changes("a").and().
		the_gate_is_entered("a").and().
		the_gate_value_becomes("a", "v2").and().
		a_subscriber_joins("late", "watch", `t('gate', 'a')`).and().
		the_gate_is_entered("a")

	when.
		the_gate_is_opened("a")

	then.
		the_watch_receives("watch", `"v2"`).and().
		the_watch_is_established("late").and().
		the_initial_watch_output_is("late", `"v2"`)
}
