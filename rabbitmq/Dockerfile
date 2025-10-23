FROM rabbitmq:4-management

# Копируем плагин
COPY ../plugins/rabbitmq_delayed_message_exchange-4.1.0.ez /opt/rabbitmq/plugins/

# Включаем плагин оффлайн
RUN rabbitmq-plugins enable --offline rabbitmq_delayed_message_exchange
