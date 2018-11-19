const http = require('http');
let Bot = require('@kikinteractive/kik');
const { Message } = Bot;
let {
	newKey,
	nextPrompt,
	answer,
} = require('./api.js');

let cfg = require('./config.json')
if (cfg.username === "") {
	throw new Error("set kik username");
}
if (cfg.apiKey === "") {
	throw new Error("set kik apiKey");
}
if (cfg.baseUrl === "") {
	throw new Error("set kik baseUrl");
}

var madlibapi = {
	key: '',
};

let bot = new Bot(cfg);
bot.updateBotConfiguration();
bot.onStartChattingMessage((message) => {
	message.reply("HEY");
});

bot.onTextMessage((message) => {
	
	// pull the user input and normalize the case
	var userInput = message.body.toLowerCase();


	if (userInput === "no") {
		message.reply("bye!")
	} else if (userInput === "yes" || madlibapi.key === '') {
		// if madlibapi.key is empty, then we need to fetch a new one

			newKey().then( (key) => {
				// save the madlib api key
				// Docs: https://madlibs.xbos.io/#apinew
				madlibapi.key = key;
				// get the information for our new madlib
				return nextPrompt(madlibapi.key);
			}).then( (resp) => {
				// 'resp' contains the information for the new madlib
				// Docs: https://madlibs.xbos.io/#apinext
				message.reply("You are doing the '" + resp.title + "' madlib");
				message.reply("Give me a " + resp.prompt);
			})
			.catch( (err) => {
				// error!
				console.error(err);
				message.reply("INTERNAL ERORR: " + err)
			});
	} else if (userInput == "skip") {
		// user wants to skip this madlib
		message.reply("Skipping!")
		newKey().then( (key) => {
							madlibapi.key = key; 
							return nextPrompt(madlibapi.key);
						}).then( (resp) => {
							message.reply("You are now doing the '" + resp.title + "' madlib");
							message.reply("Give me a " + resp.prompt);
						});
	} else {
			// user is responding to current madlib
			// Docs: https://madlibs.xbos.io/#apianswer
			answer(madlibapi.key, userInput).then( (resp) => {
				  if (resp.done) {
						message.reply("You are finished!");
						message.reply(resp.madlib);
						message.reply("Would you like to do another? (yes/no)");
						// reset api key so we can have a new madlib
						madlibapi.key = '';
					} else {
						message.reply("Give me a " + resp.prompt);
					}
				})
				.catch( (err) => {
					console.error(err);
				});
	}
});

let server = http
	.createServer(bot.incoming())
	.listen(8000, (err) => {
		if (err) {
			return console.log('something bad happened', err)
		}
		console.log('server is listening on 8000')
	});
