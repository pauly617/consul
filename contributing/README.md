# Contributing to Consul

See [our contributing guide](../.github/CONTRIBUTING.md) to get started.

This directory contains other helpful resources like checklists for making
certain common types of changes.

## Checklists

See the `checklist-*` files in this dir to see of there is a checklist that
applies to a change you wish to make. The most common one is a change to
Consul's configuration file which has a lot more steps and subtlety than might
first appear so please use the checklist!

We recommend copying the raw Markdown lists into a local file or gist while you
work and checking tasks as you complete them (by filling in `[ ]` with `[x]`).
You can then paste the list in the PR description when you post it indicating
the steps you've already done. If you want to post an initial draft PR before
the list is complete, please still include the list as is so reviewers can see
progress. You can then check the list off on the PR description directly in
GitHub's UI after pushing the relevant fixes up.

## Automation in the Repo

This repo has a mix of custom tooling, as well as several GitHub actions
stored in [.github](../.github/workflows/).

The [PR labeler](https://github.com/actions/labeler) is a common Action. This
performs labeling based on the file path that is changed. The file path
definitions are stored in [.github/pr-labeler.yml](../..github/pr-labeler.yml).

The [Issue labeler](https://github.com/github/issue-labeler) is another common
action. The label regexs are defined in
[.github/issue-labeler](../.github/issue-labeler.yml)

To fix documentation on [consul.io/docs](https://www.consul.io/docs) or
[consul.io/api-docs](https://www.consul.io/api-docs), start your branch with
`docs/`. This will skip running test suites.

`docs-cherrypick` and `backport/*` labels are put on by maintainers. To keep
documentation up to date, PRs will have this added to keep the website up to date.
## Gotchas

Some gotchas about the automation:

_Issue labeler_
- Issue labeler will have the following error in the logs: `Unexpected input(s) 'not-before', 'enable-versioned-regex', valid inputs are ['repo-token', 'configuration-path']`. This is a bug, but does not affect functionality of the actions.
- Issue labeler `jobs.triage.steps.with.not-before` is set so that old issues aren't going to churn as edits are made. This can be pushed farther and farther back to confirm tagging, but can remove labels from an issue as well. Currently there is not a way to [([disable the remove step](https://github.com/github/issue-labeler/issues/12)].
- Issue labeler claims to have [case-insensitive globbing](https://github.com/github/issue-labeler/pull/15), but I was not able to get this working.

_PR Labeler_
- This currently only acts on changed files for tags. This does not parse PR content.
- 