# Use an official Node.js runtime as a parent image
FROM node:20-alpine

# Set the working directory in the container
WORKDIR /app

# Copy package.json and install dependencies
COPY package*.json ./
RUN npm install

# Copy the rest of the application
COPY . .

# Expose the port that Next.js runs on
EXPOSE 3000

# Start the Next.js application
CMD ["npm", "run", "dev"]
