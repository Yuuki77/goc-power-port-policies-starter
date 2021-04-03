package src

const debug = false
const DefaultCommitCount = 1
const DefaultBranch = "main"

// const DefaultBranch = "factory-atlas-11907.B"
const DefaultTimeout = 30
const DefaultOutPutPath = "output"
const DefaultCsvOutPutPath = "csv"

const DefaultUrl = "http://chromium.googlesource.com/chromiumos/platform/tast-tests/"
const mainBranchLinkSelector = "body > div > div > div.RepoShortlog > div.RepoShortlog-refs > div > ul > li:nth-child(1) > a"
const commitIdSelector = "body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(1) > td:nth-child(2)"
const authorSelector = "body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(2) > td:nth-child(2)"
const commitMessageSelector = "body > div > div > pre"
const parentLinkSelector = "body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(5) > td > a"
