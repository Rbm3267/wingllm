package proxy

import (
	"fmt"
	"github.com/Rbm3267/wingllm/internal/discovery"
)

func DashboardHTML(pair *discovery.PairedModels) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>WingLLM | 2x Local AI Turbo</title>
  <style>
    :root {
      --bg: #090d16;
      --card-bg: #111827;
      --card-border: #1f2937;
      --text: #f3f4f6;
      --text-muted: #9ca3af;
      --accent: #38bdf8;
      --accent-glow: rgba(56, 189, 248, 0.15);
      --success: #34d399;
    }
    * { box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      background: var(--bg);
      color: var(--text);
      margin: 0;
      padding: 24px;
      display: flex;
      justify-content: center;
      min-height: 100vh;
    }
    .container {
      width: 100%%;
      max-width: 860px;
      display: flex;
      flex-direction: column;
      gap: 20px;
    }
    .header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      background: var(--card-bg);
      border: 1px solid var(--card-border);
      border-radius: 14px;
      padding: 16px 24px;
    }
    .header-left {
      display: flex;
      align-items: center;
      gap: 14px;
    }
    .logo-svg {
      width: 36px;
      height: 36px;
    }
    .title-group h1 {
      margin: 0;
      font-size: 20px;
      font-weight: 700;
      color: var(--text);
      display: flex;
      align-items: center;
      gap: 10px;
    }
    .title-group p {
      margin: 2px 0 0 0;
      font-size: 13px;
      color: var(--text-muted);
    }
    .badge {
      background: rgba(52, 211, 153, 0.12);
      color: var(--success);
      border: 1px solid rgba(52, 211, 153, 0.3);
      padding: 4px 10px;
      border-radius: 9999px;
      font-size: 12px;
      font-weight: 600;
    }
    .telemetry-grid {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 12px;
    }
    .stat-card {
      background: var(--card-bg);
      border: 1px solid var(--card-border);
      border-radius: 12px;
      padding: 14px 18px;
    }
    .stat-label {
      font-size: 12px;
      color: var(--text-muted);
      text-transform: uppercase;
      letter-spacing: 0.5px;
      margin-bottom: 6px;
    }
    .stat-value {
      font-size: 15px;
      font-weight: 600;
      color: var(--text);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .stat-value.accent { color: var(--accent); }
    .stat-value.speed { color: var(--success); }
    
    /* Interactive Chat Box */
    .chat-card {
      background: var(--card-bg);
      border: 1px solid var(--card-border);
      border-radius: 14px;
      display: flex;
      flex-direction: column;
      height: 480px;
      overflow: hidden;
    }
    .chat-messages {
      flex: 1;
      padding: 20px;
      overflow-y: auto;
      display: flex;
      flex-direction: column;
      gap: 14px;
    }
    .message {
      max-width: 82%%;
      padding: 12px 16px;
      border-radius: 12px;
      font-size: 14px;
      line-height: 1.5;
      white-space: pre-wrap;
    }
    .message.user {
      align-self: flex-end;
      background: #0284c7;
      color: #ffffff;
      border-bottom-right-radius: 2px;
    }
    .message.assistant {
      align-self: flex-start;
      background: #1f2937;
      color: #f3f4f6;
      border: 1px solid #374151;
      border-bottom-left-radius: 2px;
    }
    .speed-tag {
      font-size: 11px;
      color: var(--success);
      margin-top: 6px;
      font-weight: 600;
    }
    .chat-input-area {
      padding: 16px;
      background: #0d131f;
      border-top: 1px solid var(--card-border);
      display: flex;
      gap: 10px;
    }
    .chat-input {
      flex: 1;
      background: #1f2937;
      border: 1px solid #374151;
      border-radius: 8px;
      padding: 12px 14px;
      color: #fff;
      font-size: 14px;
      outline: none;
      transition: border 0.15s;
    }
    .chat-input:focus {
      border-color: var(--accent);
    }
    .send-btn {
      background: var(--accent);
      color: #0f172a;
      border: none;
      border-radius: 8px;
      padding: 0 20px;
      font-weight: 600;
      font-size: 14px;
      cursor: pointer;
      transition: opacity 0.15s;
    }
    .send-btn:hover { opacity: 0.9; }
    .send-btn:disabled { opacity: 0.4; cursor: not-allowed; }

    /* IDE Instructions */
    .info-card {
      background: var(--card-bg);
      border: 1px solid var(--card-border);
      border-radius: 12px;
      padding: 16px 20px;
      font-size: 13px;
      color: var(--text-muted);
    }
    .info-card code {
      background: #000;
      padding: 2px 6px;
      border-radius: 4px;
      color: var(--accent);
    }
  </style>
</head>
<body>
  <div class="container">
    <!-- Header -->
    <div class="header">
      <div class="header-left">
        <svg class="logo-svg" viewBox="0 0 120 120" fill="none">
          <path d="M60 16 C60.8 16 62 18.5 62.5 24 L63.5 44 L98 64 C99.5 65 99.8 66.8 98.8 67.5 L94.5 70 L63.5 61.5 L63.5 86 L74 95 C74.8 95.8 74.8 97 73.8 97.5 L70.5 98.5 L60 94 L49.5 98.5 L46.2 97.5 C45.2 97 45.2 95.8 46 95 L56.5 86 L56.5 61.5 L25.5 70 L21.2 67.5 C20.2 66.8 20.5 65 22 64 L56.5 44 L57.5 24 C58 18.5 59.2 16 60 16 Z" 
                fill="none" stroke="#38bdf8" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
        <div class="title-group">
          <h1>WingLLM <span class="badge">● 2X TURBO ACTIVE</span></h1>
          <p>Heterogeneous Speculative Decoding for Local AI</p>
        </div>
      </div>
      <div>
        <span style="font-size: 13px; color: var(--text-muted);">API: <code>http://localhost:8080/v1</code></span>
      </div>
    </div>

    <!-- Telemetry Cards -->
    <div class="telemetry-grid">
      <div class="stat-card">
        <div class="stat-label">Target Model (CPU RAM)</div>
        <div class="stat-value" id="target-model">%s</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Wingman Draft (2GB GPU)</div>
        <div class="stat-value accent" id="draft-model">%s</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Generation Speedometer</div>
        <div class="stat-value speed" id="speed-meter">Ready</div>
      </div>
    </div>

    <!-- Interactive Chat Playground -->
    <div class="chat-card">
      <div class="chat-messages" id="messages">
        <div class="message assistant">
          Hello! WingLLM is running on your machine with speculative acceleration enabled. Type a prompt below to test your local generation speed!
        </div>
      </div>
      <div class="chat-input-area">
        <input type="text" class="chat-input" id="chat-input" placeholder="Type a prompt (e.g. 'Write a fast binary search in Python')..." autofocus />
        <button class="send-btn" id="send-btn" onclick="sendMessage()">Send</button>
      </div>
    </div>

    <!-- IDE Configuration Info -->
    <div class="info-card">
      💡 <strong>Connect to VS Code (Continue / Cline)</strong>: Set Provider to <code>OpenAI</code>, Model to <code>%s</code>, and API Base URL to <code>http://localhost:8080/v1</code>.
    </div>
  </div>

  <script>
    const input = document.getElementById('chat-input');
    const sendBtn = document.getElementById('send-btn');
    const messages = document.getElementById('messages');
    const speedMeter = document.getElementById('speed-meter');

    input.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        sendMessage();
      }
    });

    async function sendMessage() {
      const text = input.value.trim();
      if (!text) return;

      input.value = '';
      input.disabled = true;
      sendBtn.disabled = true;

      // Add user message
      const userDiv = document.createElement('div');
      userDiv.className = 'message user';
      userDiv.innerText = text;
      messages.appendChild(userDiv);
      messages.scrollTop = messages.scrollHeight;

      // Add assistant placeholder
      const assistDiv = document.createElement('div');
      assistDiv.className = 'message assistant';
      assistDiv.innerText = '...';
      messages.appendChild(assistDiv);
      messages.scrollTop = messages.scrollHeight;

      speedMeter.innerText = 'Accelerating...';

      const startTime = performance.now();
      try {
        const resp = await fetch('/v1/chat/completions', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            messages: [{ role: 'user', content: text }],
            stream: false
          })
        });

        const data = await resp.json();
        const durationSec = (performance.now() - startTime) / 1000;
        const reply = data.choices && data.choices[0] ? data.choices[0].message.content : (data.content || JSON.stringify(data));
        
        // Approximate token count (words * 1.3)
        const tokenCount = data.usage ? data.usage.completion_tokens : Math.round(reply.split(/\s+/).length * 1.3);
        const tokPerSec = (tokenCount / durationSec).toFixed(2);

        assistDiv.innerText = reply;
        
        const speedTag = document.createElement('div');
        speedTag.className = 'speed-tag';
        speedTag.innerText = '⚡ ' + tokPerSec + ' tokens/sec (' + tokenCount + ' tokens in ' + durationSec.toFixed(1) + 's) [Speculative 2x Turbo]';
        assistDiv.appendChild(speedTag);

        speedMeter.innerText = tokPerSec + ' tok/s';
      } catch (err) {
        assistDiv.innerText = 'Error connecting to engine: ' + err.message;
        speedMeter.innerText = 'Error';
      } finally {
        input.disabled = false;
        sendBtn.disabled = false;
        input.focus();
        messages.scrollTop = messages.scrollHeight;
      }
    }
  </script>
</body>
</html>`, pair.Target.Name, pair.Draft.Name, pair.Target.Name)
}
