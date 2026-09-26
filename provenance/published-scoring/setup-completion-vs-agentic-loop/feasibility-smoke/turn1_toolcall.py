import json, urllib.request

preamble = """You are an agent working in a code repository. You can act by calling ONE tool per turn, using EXACTLY this format on its own line:
CALL <tool> <arg>
Tools:
  list_files <dir>      - list files in a directory
  read_file <path>      - print a file's contents
  edit_file <path>      - open a file to write (you will be prompted for content)
When you are done, write: FINAL <one-line summary>

Task: The v1.5.0 release added a ChainedFilter and fixed NullFilter empty-input handling. Tests pass, commit and tag are pushed. Finish the release.
Begin. Emit your first action now."""

wrapper = ("<|im_start|>user\n" + preamble +
           "<|im_end|>\n<|im_start|>assistant\n<think>\n\n</think>\n\n")

body = json.dumps({
  "prompt": wrapper, "n_predict": 200, "temperature": 0.8, "seed": 1,
  "top_k": 0, "top_p": 1.0, "min_p": 0.05, "repeat_penalty": 1.0, "cache_prompt": False,
}).encode()
req = urllib.request.Request("http://localhost:8081/completion", data=body,
                             headers={"Content-Type": "application/json"})
r = json.load(urllib.request.urlopen(req, timeout=60))
print("=== CONTENT ===")
print(r["content"])
print("=== stopped_limit:", r.get("stopped_limit"), "pred_n:", r.get("timings",{}).get("predicted_n"), "tok/s:", round(r.get("timings",{}).get("predicted_per_second",0),1))
