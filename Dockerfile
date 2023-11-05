FROM python:3.10-alpine
ARG USERNAME=app
ARG USER_UID_GID=10000
RUN addgroup -g $USER_UID_GID $USERNAME && adduser -u $USER_UID_GID -G $USERNAME -D $USERNAME
WORKDIR /app
COPY requirements.txt .
RUN pip install -U -r requirements.txt --no-cache-dir
COPY . .
ENV DRPP_CONTAINER_DEMOTION_UID_GID=$USER_UID_GID
CMD ["python", "main.py"]
