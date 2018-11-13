# DogStatsD Sidecar Demo

This is a basic demo on the how to set up a DogStatsD Data Dog docker sidecar in go.
We will walk through how the sidecar docker needs to be set up in compose
and then how to configure the client in your go.

## DataDog Sidecar

Here are some of the different documents online for more specifics on using the sidecar.

[DataDog Docs](https://docs.datadoghq.com/developers/dogstatsd/)

[Go Package](https://github.com/DataDog/datadog-go/tree/master/statsd)

[Go Docs](https://godoc.org/github.com/DataDog/datadog-go/statsd)

### What is The Sidecar

It is a data dog agent in the form of a docker container running along side of your services that you can send
metrics to using UDP. The sidecar will then do the aggredation and sending of those metrics to
datadog for you.

This allows you to send custom metrics to datadog in your code while it is executing and not have to
wait for the request to complete.

## The DataDog Client

This is the structure of the statsd client that we will use to communicate with the sidecar:

    type Client struct {
        writer statsdWriter

        // Namespace to prepend to all statsd calls
        Namespace stringhttps://console.aws.amazon.com/cloudwatch/home?

        // Tags are global tags to be added to every statsd call
        Tags []string

        // skipErrors turns off error passing and allows UDS to emulate UDP behaviour
        SkipErrors bool

        // BufferLength is the length of the buffer in commands.
        bufferLength int
        flushTime    time.Duration
        commands     []string
        buffer       bytes.Buffer
        stop         chan struct{}
        sync.Mutex
    }

We found that it is best to get your client by using the buffered client so that it sends
data in buffered chunks. You can do that by using the function [statsd.NewBuffered](https://godoc.org/github.com/DataDog/datadog-go/statsd#NewBuffered).

When you use the NewBuffered function it will take care of setting the endpoint and buffer length that you pass,
along with initializing the flushTime, commands, and stop variables. You should then go back and set the
namespace and tags that you need. You could use the namespace to set your projects overall name and then
tags to distinguish between items such as the aws account and environment.

### The Code Set Up

Here is an example of how to set up a client.

    ddClient, err := statsd.NewBuffered(config.Endpoint, 1)
    if err != nil {
        return nil, errors.Wrap(err, "initialize dogstatsd")
    }

    // Prefix every metric with the app name, some common tags.
    ddClient.Namespace = "sidecar-demo"
    ddClient.Tags = config.Tags

    // Use the environment to populate some more tags.
    switch config.Environment {
    case "prod":
        ddClient.Tags = append(ddClient.Tags, "account:mmcken3-demos", "environment:prod")
    case "test":
        ddClient.Tags = append(ddClient.Tags, "account:mmcken3-demos-dev", "environment:test")
    default:
        ddClient.Tags = append(ddClient.Tags, "account:"+env, "environment:"+env)
    }

Once you have this client you can then use it to start sending metrics to data dog quite easily.
Here is an example of how you would send a Gauge and and Increment:

    ddClient.Incr("task.failed", nil, 1)
    ddClient.Gauge("pipeline.lag", float64Value, nil, 1)

#### Tags

When using this sidecar to track metrics you may also want to take advantage of tags. In the code
there are a few small examples of what to do with tags in order to use them. An example of when
you want to use them is tracking the same metric like a req_timeout across two different methods,
but you need a way to tell the metrics apart.

The tag for this line sending and Incr would be `type:somethingspecial`. This param is actually a
slice so you could send in a list of tags if you want too.

    ddClient.Incr("task.happened", []string{"type:somethingspecial"}, 1)

### The Container Set Up

Here is a snippet of the docker compose set up to run the sidecar along side of your
current services.

    demo_dd_agent:
        image: datadog/agent:latest
        ports:
            - 8125:8125/udp
            - 8126:8126/tcp
        environment:
            DD_API_KEY: $DD_API_KEY

The container seems to function well in AWS Fargate with a soft memory limit of 128 MiB.
You then just need to add the env variables for `DD_API_KEY` and also set the env
variable `ECS_FARGATE=true`. There is an example json container config for Fargate in
the [file](aws-config-example.json). This is the barebones of what you would need but unless
you are doing something really special it will get the job done.

## The Demo

Ensure that you have a `.env` file with these variables set before you build and run.

    DD_API_KEY
    DEMO_NAMESPACE
    DEMO_ENVIRONMENT

Build the code with `docker-compose build` and run the code with `docker-compose up`.

You can open up another terminal window and run `docker stats` to see the containers start up.
Initially you should see one error logged out in the logs. The sidecar logs are disabled in the
demo because they can be a mess locally.

Next remove the `DD_API_KEY` from your `.env` file and re build/run the project. If you ran
`docker stats` again or noticed the missing DD_API_KEY log message, you should notice that
the sidecar container was not running. This is an important note becasue this means the messages
failed to send to the sidecar and the project kept running.

I would recomend creating some sort of `logIfError` function to wrap your calls to the sidecar in
so that you can see in your projects logs that the message has failed to send to the sidecar.

### Packages Used

[github.com/DataDog/datadog-go/statsd](https://github.com/DataDog/datadog-go/statsd)

[github.com/kelseyhightower/envconfig](https://github.com/kelseyhightower/envconfig)

[github.com/pkg/errors](https://github.com/pkg/errors)