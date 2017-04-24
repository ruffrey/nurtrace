/*
[{
    "CategoryName": "12",
    "InputText": "4+8"
}]
*/
const fs = require('fs');
const assert = require('assert');

const min = 0;
const max = 9;
const rand = () => Math.floor(Math.random() * (max - min) + min);
const desiredTotal = 10000;
const outputFile = `${__dirname}/charcat-addition.json`;
const write = () => fs.writeFileSync(outputFile, JSON.stringify(data, null, 2));
let data = [];

for (let i = 0; i < desiredTotal; i++) {
    const a = rand();
    const b = rand();
    const testCase = {
        CategoryName: `${a+b}`,
        InputText: `${a}+${b}`
    };
    const input = testCase.InputText.split('');
    assert.equal(
        parseInt(testCase.CategoryName, 10),
        parseInt(input[0], 10) + parseInt(input[2]),
        "Failed with test case " + JSON.stringify(testCase, null, 2)
    );
    data.push(testCase);
}

write();
