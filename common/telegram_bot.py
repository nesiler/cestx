import logging
import os
from flask import Flask, jsonify, request
from dotenv import load_dotenv
from telegram import InlineKeyboardButton, InlineKeyboardMarkup, Update
from telegram.ext import Application, CallbackQueryHandler, CommandHandler, ContextTypes, MessageHandler, filters
from threading import Thread
import asyncio  # Import asyncio for subprocess handling

# Load environment variables from .env file
load_dotenv()
TELEGRAM_TOKEN = os.getenv("TELEGRAM_TOKEN")
CHAT_ID = os.getenv("CHAT_ID")

app = Flask(__name__)

# Enable logging
logging.basicConfig(
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s", level=logging.INFO
)

# set higher logging level for httpx to avoid all GET and POST requests being logged
logging.getLogger("httpx").setLevel(logging.WARNING)

logger = logging.getLogger(__name__)

async def start(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    keyboard = [
        [InlineKeyboardButton("Deploy", callback_data="1")],
        [InlineKeyboardButton("Status", callback_data="2")],
    ]

    reply_markup = InlineKeyboardMarkup(keyboard)

    await update.message.reply_text("Please choose:", reply_markup=reply_markup)


async def button(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    """Parses the CallbackQuery and updates the message text."""
    query = update.callback_query
    await query.answer()
    
    if query.data == "1":
        await deployer(update, context)
    elif query.data == "2":
        await status(update, context)


async def deployer(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    # Restart starter.service
    await sender(message="!!! SYSTEM RESTARTING!!!")
    os.system("systemctl restart starter.service")

    await asyncio.sleep(10)
    
    proc = await asyncio.create_subprocess_shell(
        "systemctl status deployer.service",
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.PIPE
    )

    stdout, stderr = await proc.communicate()
    
    # Check if there was an error getting the status
    if stderr:
        error_message = f"!!! Error checking deployer.service status:\n{stderr.decode()} !!!"
        await sender(message=error_message)
        return

    # Send the status message
    status_message = stdout.decode()

    # Optionally, filter the status output for brevity (if needed)
    filtered_status = "\n".join(status_message.split("\n")[:5])  # Get first 5 lines
    await sender(message=f"!!! deployer.service status:\n`\n{filtered_status}\n` !!!")


async def status(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    await sender(message="System status: ...")
    

async def sender(message: str) -> None:
    application = Application.builder().token(TELEGRAM_TOKEN).build()
    await application.bot.send_message(chat_id=CHAT_ID, text=message)

@app.route('/send', methods=['POST'])
async def send_message():  
    data = request.get_json()
    if not data or 'message' not in data:
        return jsonify({"error": "Invalid request data"}), 400
    message = data['message']
    try:
        await sender(message)
        return jsonify({"status": "Message sent successfully"})
    except Exception as e:
        logger.error(f"Error sending message: {e}")
        return jsonify({"error": "Failed to send message"}), 500
    


def main() -> None:
    # Create the Application and pass it your bot's token.
    application = Application.builder().token(TELEGRAM_TOKEN).build()
    job_queue = application.job_queue

    application.add_handler(CommandHandler("start", start))
    application.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, start))
    application.add_handler(CallbackQueryHandler(button))

    # Run Flask and Telegram bot in separate threads
    flask_thread = Thread(target=app.run, kwargs={'host': '0.0.0.0', 'port': 5005})
    flask_thread.start()

    application.run_polling(allowed_updates=Update.ALL_TYPES)

if __name__ == "__main__":
    main()