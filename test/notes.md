mosquitto_pub -i bitflow-engine-4-registry-request-publisher -t 'bitflow/engine/4/registry-request' -m "register"
# start stop
mosquitto_pub -i bitflow-engine-4-registry-request-publisher -t 'bitflow/engine/4/registry-request' -m "deregister"
# event publishing

