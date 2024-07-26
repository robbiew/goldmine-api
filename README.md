# Gold Mine Log Stats API

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
- Synchronet log files in the specified directory (e.g. /var/log)

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/robbiew/goldmine-api.git
   cd synchronet-log-stats-api```

2.	Build the project:
    ```bash
    go build -o goldmine-api main.go```

### Usage

1. Run the server:
   ```bash
   sudo ./goldmine-api --logdir=/path/to/your/log/dir```

Note, for Linux Synchronet defaults logs to /var/log/syslog*. By default, the server listens on port 8080.

2. Access the API endpoints:
* Top 10 Games: http://localhost:8080/top10?period=all
  * Replace all with month or year or a specific month (e.g., july) or year (e.g., 2024).
* Detailed Stats: http://localhost:8080/stats?period=all
  * Replace all with month or year or a specific month (e.g., july) or year (e.g., 2024).
 
### API Endpoints

GET /top10
* Retrieve the top 10 most launched games.
  
Query Parameters:
* period (required): The time period for the statistics. Valid values are month, year, all, or a specific month (e.g., july) or year (e.g., 2024).
  
Response:
* 200 OK: A JSON object containing the top 10 most launched games.
* 400 Bad Request: If the period parameter is missing or invalid.

```{
  "period": "all",
  "games": [
    {
      "game_name": "Adventurer's Maze II",
      "launch_count": 42
    },
    ...
  ]
}
```
GET /stats
* Retrieve detailed statistics.
  
Query Parameters:
* period (required): The time period for the statistics. Valid values are month, year, all, or a specific month (e.g., july) or year (e.g., 2024).

Response:
* 200 OK: A JSON object containing detailed statistics.
* 400 Bad Request: If the period parameter is missing or invalid.

```
{
  "month": {
    "january": [
      {
        "game_name": "Adventurer's Maze II",
        "launch_count": 42
      },
      ...
    ],
    ...
  }
}
```


