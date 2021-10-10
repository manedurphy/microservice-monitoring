#! /bin/bash
ITERATIONS=$1
METHODS=$2

docker run -it --network="tracing" \
	-e JAEGER_SERVICE_NAME=client \
	-e JAEGER_AGENT_HOST=jaeger \
	-e JAEGER_AGENT_PORT=6831 \
	-e JAEGER_SAMPLER_MANAGER_HOST_PORT=jaeger:6831 \
	-e JAEGER_REPORTER_LOG_SPANS=true \
	-e JAEGER_SAMPLER_PARAM=1 \
    -e JAEGER_SAMPLER_TYPE=const \
	jaeger-client --iterations="$ITERATIONS" --methods="$METHODS"
