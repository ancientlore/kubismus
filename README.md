![logo](https://raw.githubusercontent.com/ancientlore/kubismus/master/media/kubismus36.png) kubismus
====================================================================================================

[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ancientlore/kubismus?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Build Status](https://travis-ci.org/ancientlore/kubismus.svg?branch=master)](https://travis-ci.org/ancientlore/kubismus)
[![GoDoc](https://godoc.org/github.com/ancientlore/kubismus?status.svg)](https://godoc.org/github.com/ancientlore/kubismus)

Kubismus is a [go](http://golang.org/) package that makes it easy to display status metrics using [Cubism.js](https://square.github.io/cubism/). ("Kubismus" is the German word for "cubism".)

Checkout the sample code: [ktest.go](https://gist.github.com/ancientlore/7802445bba933a2c85a8)

Screen Shot
-----------

![screenshot](https://raw.githubusercontent.com/ancientlore/kubismus/master/media/screenshot.jpg)

Defining the HTTP Handler
-------------------------

To get started with all defaults, simply register the HTTP handler and serve HTTP:

	kubismus.HandleHTTP()
	go http.ListenAndServe(":8080", nil)

This creates an endpoint at http://localhost:8080/kubismus that will register information you log.

If you need a custom endpoint, use `kubismus.ServeHTTP` directly:

	http.Handle("/", http.HandlerFunc(kubismus.ServeHTTP))

Adding Data
-----------

Kubismus shows graphs and a table of data. You can add entries to these at any time. To add an entry to the table that shows the number of goroutines:

	kubismus.Note("Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine()))

To add an entry to a graph:

	kubismus.Metric("Metric Name", count, value)

By default, each metric has a count, average, and sum graph. To configure which graphs to show, use the `Define` method:

	kubismus.Define("Posts", kubismus.COUNT, "HTTP Posts")
	kubismus.Define("Posts", kubismus.SUM, "Bytes Posted")

Adding metrics and table entries use channels and are thread-safe. Graphs for new metrics may not appear until a browser refresh.

Customizing the Title
---------------------

You can configure the status page's icon and title:

	kubismus.Setup("My Cool Utility", "/web/kubismus36.png")
