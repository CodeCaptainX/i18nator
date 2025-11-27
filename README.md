# i18nator

i18nator is a CLI tool that manages i18n JSON files with automatic translation support for multiple languages.

## Installation

### 1. Clone the Repository

```bash
git clone <repository_url>
cd <repository_folder>
```

### 2. Build the Binary

```bash
go build -o i18nator
```

### 3. Move the Binary to `/usr/local/bin`

```bash
sudo mv i18nator /usr/local/bin
```

## Usage

### Add a new translation key

```bash
i18nator add <key> "<value>"
```

Example:
```bash
i18nator add errors.card.expired "Your card has expired"
```

This will create the key in all language files (`en.json`, `km.json`, `zh.json`) with automatic translation.

### List all translation keys

```bash
i18nator list
```

### Update an existing key

```bash
i18nator update <key> "<new_value>"
```

Example:
```bash
i18nator update errors.card.expired "Your card is expired"
```

### Remove a translation key

```bash
i18nator remove <key>
```

Example:
```bash
i18nator remove errors.card.expired
```

## Features

- ✅ Automatic translation to Khmer and Chinese
- ✅ Manages JSON files in `pkg/translates/localize/i18n`
- ✅ Creates files automatically if they don't exist
- ✅ Sorted JSON output for clean diffs

## License

This project is licensed under the MIT License.

## Contributing

Feel free to submit issues and pull requests to improve this project.