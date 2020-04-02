In response to [this challenge](https://gist.github.com/mitechie/c904221e06dd3a2f4158938e256be23c), I threw together a small daemon that uses goroutines and channels to ingest metrics from a [dummy API](https://github.com/juju/demoware) and dispatch them to payload-specific handlers.

Assumptions/clarifications:
* Constant CPU count for all requests. Set on the first metric ingested, errors on subsequent requests with different CPU counts.
* The requirement "keep track of the most recent timestamp" means comparing timestamps rather than just storing the timestamp that was received most recently. Even though in the case of this demo, both would behave the same since [the demoware API increases the timestamp monotonically](https://github.com/juju/demoware/blob/master/main.go#L206)

I completed the core functionality in about 4 hours, polished everything for another hour. Didn't get to the following in the recommended time:
* `Handling of back-pressure: match the data ingestion component's API poll rate to the processing speed of downstream components.`
    * I imagine I'd utilize `context`s or buffered channels for the first, but would need to investigate more.
* `Per-component introspection endpoints so we can query/monitor their internal state.`
    *Prometheus would probably be best if that was available to the project already. Otherwise, I thought of having each handler return a channel of channels, so that callers could send a channel to the handler and the handler would send back the current stats over that, but then it sounded like it could get hairy with writing to a potentially closed channel. Again, would just need more time to evaluate options.
* `Integration/end-to-end testing`
    * This would have been made easier once the per-component introspection was introduced, so I was putting it off thinking I'd have time to do it.
* Configuration management / commandline flags
    * Since its just a demo, I didn't bother. Would probably use [viper](https://github.com/spf13/viper) if making this more production ready. Then the basic authentication provided by the demoware API could be utilized.
* Exponential backoff on data ingestion request errors
    * Again, more of a prod consideration.

Perhaps I'll come back to the first two items later since they seem like interesting learning experiences.

Footnotes:
The concurrency patterns found in `metrics/helpers.go` are from the book [Concurrency in Go](http://shop.oreilly.com/product/0636920046189.do) by Katherine Cox-Buday. They're mostly to enhance readability.