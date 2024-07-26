# Synchronet Log Stats API

This project provides an HTTP API to generate and serve statistics from Synchronet log files for the [Gold Mine Game Server](http://goldminebbs.com). The statistics include the top 10 most launched games and detailed stats based on monthly, yearly, or all-time data.

## Features

- Parses log files from a specified directory.
- Generates JSON statistics for:
  - Top 10 most launched games.
  - Monthly statistics.
  - Yearly statistics.
  - All-time statistics.
- Automatically reloads and refreshes data every 24 hours.

## Getting Started

### Prerequisites

- Go 1.14 or later
- Synchronet log files in the specified directory

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/synchronet-log-stats-api.git
   cd synchronet-log-stats-api```

2.	Build the project:
