const sigma = require('sigma');
global.sigma = sigma;
require('./forceAtlas2/worker');
require('./forceAtlas2/supervisor');
const gunzip = require('zlib').gunzipSync
const fs = require('fs');
const rawText = gunzip(fs.readFileSync("../net.nur"));
const nw = JSON.parse(rawText.toString('utf8'));
const g = {
    nodes: [],
    edges: []
};
console.log({ nw })
Object.keys(nw.Cells).forEach(cellId => {
    const cell = nw.Cells[cellId];
    g.nodes.push({
        id: cell.ID,
        label: cell.Tag || cell.ID,
        size: Object.keys(cell.AxonSynapses).length + Object.keys(cell.DendriteSynapses).length,
        color: '#' + (Math.floor(Math.random() * 16777215).toString(16) + '000000').substr(0, 6)
    });
});
Object.keys(nw.Synapses).forEach(synapseId => {
    const synapse = nw.Synapses[synapseId];
    g.edges.push({
        id: synapse.ID,
        source: synapse.FromNeuronAxon,
        target: synapse.ToNeuronDendrite
    });
});
console.log({ g });
// const i,
//     s,
//     o,
//     N = 1000,
//     E = 5000,
//     C = 5,
//     d = 0.5,
//     cs = [];
// // Generate the graph:
// for (i = 0; i < C; i++)
//     cs.push({
//         id: i,
//         nodes: [],
//         color: '#' + (
//             Math.floor(Math.random() * 16777215).toString(16) + '000000'
//         ).substr(0, 6)
//     });
// for (i = 0; i < N; i++) {
//     o = cs[(Math.random() * C) | 0];
//     g.nodes.push({
//         id: 'n' + i,
//         label: 'Node' + i,
//         x: 100 * Math.cos(2 * i * Math.PI / N),
//         y: 100 * Math.sin(2 * i * Math.PI / N),
//         size: Math.random(),
//         color: o.color
//     });
//     o.nodes.push('n' + i);
// }
// for (i = 0; i < E; i++) {
//     if (Math.random() < 1 - d)
//         g.edges.push({
//             id: 'e' + i,
//             source: 'n' + ((Math.random() * N) | 0),
//             target: 'n' + ((Math.random() * N) | 0)
//         });
//     else {
//         o = cs[(Math.random() * C) | 0]
//         g.edges.push({
//             id: 'e' + i,
//             source: o.nodes[(Math.random() * o.nodes.length) | 0],
//             target: o.nodes[(Math.random() * o.nodes.length) | 0]
//         });
//     }
// }
s = new sigma({
    graph: g,
    container: 'graph-container',
    settings: {
        drawEdges: true
    }
});
// Start the ForceAtlas2 algorithm:
s.startForceAtlas2({
    worker: false,
    barnesHutOptimize: false
});
