for system in redis rabbitmq postgresql; do
    ansible-playbook -i inventory.yaml launch.yaml --extra-vars "target=$system"
done