FROM mcr.microsoft.com/devcontainers/javascript-node:20-bookworm

RUN npm install -g @angular/cli@17

WORKDIR /workspace

COPY package*.json ./
COPY Makefile ./

RUN npm install

COPY . .

EXPOSE 3000 4000
