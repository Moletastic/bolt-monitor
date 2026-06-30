## Purpose

Define the visual iconography used to identify notification channel types on dashboard surfaces.

## ADDED Requirements

### Requirement: System renders an icon per notification channel type

The dashboard SHALL render a recognizable icon next to each notification channel type label on operator surfaces that list or summarize channels.

#### Scenario: Operator scans channels list

- **WHEN** operator opens the notification channels list
- **THEN** each channel row shows a small icon next to the channel type label
- **AND** the icon is consistent with the channel type semantic (delivery method)

#### Scenario: Channel type maps to a known icon

- **WHEN** the dashboard renders a channel type icon
- **THEN** `telegram` renders the send glyph
- **AND** `email` renders the mail glyph
- **AND** `sms` renders the message glyph
- **AND** `webhook` renders the webhook glyph
- **AND** `pagerduty` renders the siren glyph

#### Scenario: Icon appears next to the text label

- **WHEN** the dashboard displays the channel type in any list or detail surface
- **THEN** the type label text remains present alongside the icon
- **AND** the icon does not replace the readable type label
