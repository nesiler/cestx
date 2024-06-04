FROM python:3.9-slim-buster
LABEL authors="nesiler"

# Install required packages
RUN pip3 install --no-cache-dir "python-telegram-bot[all]" "flask[async]" python-dotenv

# Set the working directory in the container
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . /app

# Expose the port Flask will run on
EXPOSE 5005

# Define the command to run your application
CMD ["python3", "telegram_bot.py"]