# Google Gemini in Discord

This project use Google Gemini API to realize chatting with generative model in
Discord. Each chat session can be created in a **thread**


## Dev Locally

1. Prepare a configuration file named `config.yaml` with the following content

```yaml
firebase:
  projectId: {YOUR_FIREBASE_PROJECT_ID}
gemini:
  apiKey: {YOUR_GEMINI_API_KEY}
discord:
  botToken: {YOUR_DISCORD_BOT_TOKEN}
```

- for `firebase.projectId`, which is used to store chat session in firestore
- for `gemini.apiKey`, which is used to call Gemini API with prompt and generate
  response
- for `discord.botToken`, which is used to authenticate Discord bot


## Chat Session Storage

There are 2 different ways to store chat session

1. store in memory

    In file `repository/chat_session/memory_chat_sess_repository`, I used a struct
    to store chat session. but this will be reset each time the server restart.

1. store in firebase `firestore` (to be implemented)

    firestore is a document-based NoSQL data storage. [ref](https://firebase.google.com/docs/firestore)

