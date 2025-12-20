FROM node:22.19.0 as build

ARG API_ENDPOINT
ENV VITE_API_ENDPOINT=$API_ENDPOINT

ARG INSECURE 
ENV VITE_INSECURE=$INSECURE

WORKDIR /app

COPY . .

RUN npm install && npm run build

FROM nginx:1.29.1

COPY --from=build /app/dist /usr/share/nginx/html
COPY ./nginx.conf ./etc/nginx/nginx.conf

CMD ["nginx", "-g", "daemon off;"]
