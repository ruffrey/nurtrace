/*
4+8=12
*/
const fs = require('fs');
const assert = require('assert');

const min = 0;
const max = 9;
const rand = () => Math.floor(Math.random() * (max - min) + min);
const desiredTotal = 10000;
const outputFile = `${__dirname}/charcat-addition.txt`;
const write = () => fs.writeFileSync(outputFile, data);
let data = '';

for (let i = 0; i < desiredTotal; i++) {
    const a = rand();
    const b = rand();
    const nextItem = `${a}+${b}=${a+b}\n`;
    if (!data.includes(nextItem)) {
        data += nextItem;
    }
}

write();
