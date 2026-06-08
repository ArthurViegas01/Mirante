# syntax=docker/dockerfile:1

# ---- build ----
FROM node:24-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

# ---- run (non-root) ----
FROM node:24-alpine AS run
WORKDIR /app
ENV NODE_ENV=production
ENV PORT=3000
COPY --from=build /app/build ./build
COPY --from=build /app/package.json ./package.json
USER node
EXPOSE 3000
CMD ["node", "build"]
