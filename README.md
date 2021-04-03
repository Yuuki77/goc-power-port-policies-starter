
## About The Project

This repo is for implementation of for a starter bug of GSoC 2021(Port power policies end-to-end tests project.)

See more detail [here](https://docs.google.com/document/d/1mAPQ1vpnFdiKo89oyLOPkM87bDW1WUVC4NWZC1mNiYY/edit?hl=en&forcehl=1)

<!-- GETTING STARTED -->
## Getting Started

### Prerequisites
Please make sure you have installed golang in your environments
### Installation

1. Run chrome os with debug mode
For example, in Mac ` /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222`
2. Clone the repo
3. `$ go run main.go`


<!-- USAGE EXAMPLES -->
## Usage

It does
1. Navigate to chromium.googlesource.com/chromiumos/platform/tast-tests/
2. Click main under branches on the left, it navigates you to the latest commit
3. Parse page and write only commit message to the file. By default it will create a folder called output
4. Click on the commit hash on the right side of parent.
5. Generate csv file to show how many commits each contributor created (author) and reviewed (Reviewed-by).
By Default it will create a folder called csv


Also you can pass some parameters
1. Commit Numbers by default `1`.  
  You can specify by --numbers
2. Repository URL by default `http://chromium.googlesource.com/chromiumos/platform/tast-tests/`  
  You can specify by --url
3. Branch name by default `main`.  
  You can specify by --branch
4. Timeout in sec --timeout by default 30 sec.  
  You can specify by --timeout
5. Commit message files folder path by default `output`.  
  You can specify by --output

6. CSV files folder path. By default `csv`.  
You can specify by --csv-output

