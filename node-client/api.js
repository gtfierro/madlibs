const request = require('superagent');

const API_URL = "https://madlibs.xbos.io"

async function newKey() {
	let res = await request.get(API_URL + "/api/new");
	console.log("New Key: " + res.body.key);
	return res.body.key;
}

async function nextPrompt(key) {
	console.log('key',key)
	let res = await request.post(API_URL + "/api/next").send({ key: key });
	return res.body;
}

async function answer(key, answer) {
	console.log('key',key,'answer',answer)
	await request.post(API_URL + "/api/answer")
	    				.set('Content-Type', 'application/json')
							.send({ key: key, answer: answer});
	return nextPrompt(key);
}

module.exports = {
	newKey,
	nextPrompt,
	answer,
}
