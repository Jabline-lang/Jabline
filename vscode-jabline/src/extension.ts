import * as vscode from 'vscode';
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  StreamInfo,
} from 'vscode-languageclient/node';
import { spawn, ChildProcess } from 'child_process';

let client: LanguageClient;
let lspProcess: ChildProcess | null = null;

function getJablineConfig(): { exePath: string; sandbox: string; lspEnabled: boolean } {
  const config = vscode.workspace.getConfiguration('jabline');
  return {
    exePath: config.get<string>('executablePath', 'jabline'),
    sandbox: config.get<string>('sandboxLevel', 'none'),
    lspEnabled: config.get<boolean>('lsp.enabled', true),
  };
}

function startLspClient(context: vscode.ExtensionContext): LanguageClient | null {
  const { exePath, lspEnabled } = getJablineConfig();
  if (!lspEnabled) return null;

  const serverOptions: ServerOptions = () => {
    lspProcess = spawn(exePath, ['lsp'], {
      stdio: ['pipe', 'pipe', 'pipe'],
    });

    lspProcess.on('error', (err: Error) => {
      vscode.window.showErrorMessage(
        `Failed to start Jabline LSP: ${err.message}. Ensure '${exePath}' is available in PATH.`
      );
    });

    lspProcess.on('exit', (code: number | null) => {
      if (code !== 0) {
        console.error(`Jabline LSP exited with code ${code}`);
      }
      lspProcess = null;
    });

    const streamInfo: StreamInfo = {
      writer: lspProcess!.stdin!,
      reader: lspProcess!.stdout!,
    };
    return Promise.resolve(streamInfo);
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [{ scheme: 'file', language: 'jabline' }],
    synchronize: {
      fileEvents: vscode.workspace.createFileSystemWatcher('**/*.jb'),
    },
    traceOutputChannel: vscode.window.createOutputChannel('Jabline LSP Trace'),
    initializationOptions: {},
  };

  const client = new LanguageClient('jabline', 'Jabline Language Server', serverOptions, clientOptions);
  return client;
}

function runJablineFile(sandbox: string) {
  const editor = vscode.window.activeTextEditor;
  if (!editor) {
    vscode.window.showErrorMessage('No active editor');
    return;
  }

  const document = editor.document;
  if (document.languageId !== 'jabline') {
    vscode.window.showErrorMessage('Active file is not a Jabline file (.jb)');
    return;
  }

  const filePath = document.fileName;
  const { exePath } = getJablineConfig();

  const terminal = vscode.window.createTerminal({
    name: 'Jabline Run',
    hideFromUser: false,
  });
  terminal.show();

  // Save before running
  document.save().then(() => {
    const sandboxFlag = sandbox !== 'none' ? ` --sandbox ${sandbox}` : '';
    terminal.sendText(`${exePath} run${sandboxFlag} "${filePath}"`);
  });
}

function buildJablineFile(sandbox: string) {
  const editor = vscode.window.activeTextEditor;
  if (!editor) {
    vscode.window.showErrorMessage('No active editor');
    return;
  }

  const document = editor.document;
  if (document.languageId !== 'jabline') {
    vscode.window.showErrorMessage('Active file is not a Jabline file (.jb)');
    return;
  }

  const filePath = document.fileName;
  const { exePath } = getJablineConfig();

  document.save().then(() => {
    const outputPath = filePath.replace(/\.jb$/, '');
    const sandboxFlag = sandbox !== 'none' ? ` --sandbox ${sandbox}` : '';

    const terminal = vscode.window.createTerminal({
      name: 'Jabline Build',
      hideFromUser: false,
    });
    terminal.show();
    terminal.sendText(`${exePath} build${sandboxFlag} -o "${outputPath}.exe" "${filePath}"`);
  });
}

export function activate(context: vscode.ExtensionContext) {
  // Register commands
  context.subscriptions.push(
    vscode.commands.registerCommand('jabline.runFile', () => {
      const { sandbox } = getJablineConfig();
      runJablineFile(sandbox);
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('jabline.buildFile', () => {
      const { sandbox } = getJablineConfig();
      buildJablineFile(sandbox);
    })
  );

  // Start LSP client
  client = startLspClient(context)!;
  if (client) {
    context.subscriptions.push(client);
    client.start().then(() => {}, () => {});
  }
}

export function deactivate(): Thenable<void> | undefined {
  if (lspProcess) {
    lspProcess.kill();
    lspProcess = null;
  }
  if (client) {
    return client.stop();
  }
  return undefined;
}
