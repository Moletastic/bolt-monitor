## ADDED Requirements

### Requirement: Telegram sender sends messages via Telegram Bot API
The system SHALL send messages using `https://api.telegram.org/bot<token>/sendMessage`.

#### Scenario: Send notification message
- **WHEN** `TelegramSender.Send()` is called with a valid bot token and chat ID
- **THEN** it POSTs to Telegram Bot API with `chat_id` and `text` parameters
- **AND** returns nil on success

### Requirement: Telegram sender validates config before sending
The system SHALL return an error if `botToken` or `chatId` is empty when `Send()` is called.

#### Scenario: Missing bot token
- **WHEN** `TelegramSender.Send()` is called with empty bot token
- **THEN** it returns an error indicating bot token is required

#### Scenario: Missing chat ID
- **WHEN** `TelegramSender.Send()` is called with empty chat ID
- **THEN** it returns an error indicating chat ID is required

### Requirement: Telegram sender implements NotificationSender interface
The system SHALL have `TelegramSender` implement the `NotificationSender` interface with `Send()`, `ChannelType()`, and `ValidateConfig()`.

#### Scenario: ChannelType returns telegram
- **WHEN** `TelegramSender.ChannelType()` is called
- **THEN** it returns the string `telegram`

#### Scenario: ValidateConfig with valid config
- **WHEN** `TelegramSender.ValidateConfig()` is called with valid bot token
- **THEN** it returns nil

#### Scenario: ValidateConfig with invalid config
- **WHEN** `TelegramSender.ValidateConfig()` is called with empty bot token
- **THEN** it returns a validation error

### Requirement: Telegram sender escapes special characters
The system SHALL escape special characters in the message text for Telegram MarkdownV2 compatibility.

#### Scenario: Message with special characters
- **WHEN** `TelegramSender.Send()` receives a message with characters like `_`, `*`, `[`
- **THEN** it escapes them appropriately for Telegram MarkdownV2 format

### Requirement: Telegram sender supports silent delivery
The system SHALL support `disable_notification` flag in the API call when configured.

#### Scenario: Send silent notification
- **WHEN** channel config has `sendSilently: true`
- **THEN** the Telegram API call includes `disable_notification: true`
