# Nurtrace CLI tool: `nt`


## Usage

Training:

```bash
nt train -n network.nur -d ../data/iris.json -v vocab.json
```

Testing / evalating:

```bash
nt sample -v vocab.json --seed=5.0,3.2,1.2,0.2 network.nur
```