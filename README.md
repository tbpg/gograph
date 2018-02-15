GoGraph
=======

GoGraph builds graphs out of Go source code.

GoGraph currently only works with Structs.

Install
-------

```bash
go get -u github.com/tbpg/gograph
```

Sample
------

```bash
gograph gonum.org/v1/gonum/graph/simple.DirectedGraph
dot -Tpng out.dot -o out.png
```

![sample graph](./sample.png)

Questions
---------

File a bug or reach out on [Twitter](http://twitter.com/tbpalsulich).

Disclaimer
----------

This is not an official Google product.