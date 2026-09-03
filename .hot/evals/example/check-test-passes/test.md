+++
base = "go-project"
intent = "Add a function that returns 42 and write a test for it"

[[graders]]
type = "file_exists"
path = "main.go"

[[graders]]
type = "git_dirty"
min = 1
max = 5
+++

Agent should modify the project by adding a function and a test file.
The working tree should be dirty after the agent finishes.
