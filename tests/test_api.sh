#!/usr/bin/env bash
host="http://localhost"
kafka_broker="http://$LOCACL_IP:9094"
register_endpoint="auth/register"
login_endpoint="auth/login"
order_endpoint="orders/order"
order_list_endpoint="orders/list"
kafka_orders_topic="new-order"

RED='\033[0;31m'    # Red color
GREEN='\033[0;32m'  # Green color
NC='\033[0m'        # No color
token=""

login_postfix=$((1 + $RANDOM % 10000))
testsResult=()
source .env

register_login() {
    response=$(curl -s -X POST "$host/$1" \
        -H "Content-Type: application/json" \
        -d "{\"login\": \"user-$login_postfix\", \"pwd\": \"$2\"}")

    exit_code=$?
    if [ $exit_code -eq 0 ]; then
        echo "Request was successful."
        echo "Response: $response"
        status=$(echo "$response" | jq -r '.status')
        if [ "$status" = "ok" ]; then
            echo -e "$3"
            testsResult+=("$3")
            token=$(echo "$response" | jq -r '.token')
        else 
            echo -e "$4"
            testsResult+=("$4")
        fi
    else
        echo "Request failed with exit code: $exit_code"
        testsResult+=("${RED}Registration request failed${NC}")
    fi
}

check_kafka_message() {
    order_key=$1
    consumer_output=$($LOCAL_KAFKA_PATH/kafka-console-consumer.sh --bootstrap-server $kafka_broker --topic $kafka_orders_topic --from-beginning --timeout-ms 5000 --property print.key=true --property key.separator=',')
    echo $consumer_output 
    if echo "$consumer_output" | grep -q "^$order_key,"; then
        echo -e "${GREEN}OK: Order key found in Kafka${NC}"
        testsResult+=("${GREEN}OK: Order key found in Kafka${NC}") 
    else
        echo -e "${RED}FAILED: Order key not found in Kafka${NC}"
        testsResult+=("${RED}FAILED: Order key not found in Kafka${NC}")  #
    fi
}

# 1. Simple registration
register_login $register_endpoint "pwd123" "${GREEN}OK: Registration succeeded${NC}" "${RED}FAILED: Registration failed${NC}" 

# 2. Registration with same login
register_login $register_endpoint "pwd123" "${RED}FAILED: Registration succeeded with same login${NC}" "${GREEN}OK: Registration rejected (same login)${NC}"

# 3. Login with wrong pwd
register_login $login_endpoint "fff" "${RED}FAILED: Registration succeeded with wrong pwd${NC}" "${GREEN}OK: Login rejected (wrong creds)${NC}"

# 4. Login
register_login $login_endpoint "pwd123" "${GREEN}OK: Login succeeded${NC}" "${RED}FAILED: Login failed${NC}" 


# 5. Order items
order_response=$(curl -s -X POST $host/$order_endpoint -H "Content-Type: application/json" -H "Authorization: $token" -d '{"items": {"1": 2, "2": 3, "4": 10}, "delivery_addr": "123 Main Sfdnsiofnmdskt"}')
echo $order_response
status=$(echo "$order_response" | jq -r '.status')
if [ "$status" = "ok" ]; then
    echo -e "${GREEN}OK: Ordered succeeded${NC}"
    testsResult+=("${GREEN}OK: Ordered succeeded${NC}")
    order_id=$(echo "$order_response" | jq -r '.order_id')

    # 6. Check kafka for new order 
    check_kafka_message $order_id

    # 7. Order list
    order_list_response=$(curl -s $host/$order_list_endpoint -H "Authorization: $token")
    echo $order_list_response
    echo  $order_id
    if echo "$order_list_response" | jq -e ".[] | select(.order_id == $order_id)" > /dev/null; then
        echo -e "${GREEN}OK: Order ID $order_id found in order list${NC}"
        testsResult+=("${GREEN}OK: Order ID $order_id found in order list${NC}")
    else
        echo -e "${RED}FAILED: Order ID $order_id not found in order list${NC}"
        testsResult+=("${RED}FAILED: Order ID $order_id not found in order list${NC}")
    fi
else 
    echo -e "${RED}FAILED: Order failed${NC}"
    testsResult+=("${RED}FAILED: Order failed${NC}")
    
fi



printf "\n\nTests result\n"
for t in "${testsResult[@]}"; do
  echo -e "$t"
done