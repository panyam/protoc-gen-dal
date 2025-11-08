
Please thoroughly read the SUMMARY.md on design choices and tradeoffs and our overall vision and requirements.

Then read the TODO.md for a list of what is to be done.  Id suggest start with the next uncompleted item in this list.

When developing features always:

* Write a failing test first before adding/modifying/refactoring code and then make it work to ensure all tests pass.
* Then checkpoint by:
  - Updating SUMMARY.md if any of our design decisions, tradeoffs and requirements have changed
  - Updating TODO.md on updated progress as well as adding new items to be done and removing items that are no longer
    relevant or needed.
* Then commit

Few available commands:

`make buf`: For generating protos from spec
`make test`: Runs ALL tests if you want to do it in one go.
`make build`: Builds all binaries and puts them in the ./bin folder so you can access them from there and only there

A few things about style:
* When you are writing functions - explain what is is meant to do and how it fits in the larger picture in its comments (Do not worry about trivial details).
* When you modify functions see if the comments also needs an update on its purpose and behavior.
* When writing tests see if you can refactor common parts/utilities and minimize over duplications
* When writing documentation avoid emojis (except for check boxes on todo lists etc).  Also avoid flowerly language full
  of superfluous adjectives.
