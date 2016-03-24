# structmatcher
Query language for matching structures

[![Build Status](https://secure.travis-ci.org/cgrates/structmatcher.png)](http://travis-ci.org/cgrates/structmatcher)

The StructMatcher type will parse a condition string and match it against a given structure.

The condition syntax is a json encoded string similar to mongodb query language.

Examples:
- {"Weight":{"*gt":50}} checks for a balance with weight greater than 50
- {"*or":[{"Value":{"*eq":0}},{"Value":{"*gte":100}}] checks for a balance with value equal to 0 or equal or highr than 100

Available operators:
- *eq: equal
- *gt: greater than
- *gte: greater or equal than
- *lt: less then
- *lte: less or equal than
- *exp: expired
- *or: logical or
- *and: logical and
- *has: receives a list of elements and checks that the elements are present in the specified field (StringMap type)

Equal (*eq) and local and (*and) operators are implicit for shortcuts. In this way:

{"*and":[{"Value":{"*eq":3}},{"Weight":{"*eq":10}}]} is equivalent to: {"Value":3, "Weight":10}.
