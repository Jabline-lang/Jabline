const path = require('path');
const { workspace, window } = require('vscode');
const { LanguageClient, TransportKind } = require('vscode-languageclient/node');

let client;

function activate(context) {
    const config = workspace.getConfiguration('jabline');

    // 1. Use explicit setting if provided
    let command = config.get('executablePath');

    // 2. If no setting, try to find jabline.exe in the workspace root
    if (!command || command === 'jabline') {
        const workspaceFolders = workspace.workspaceFolders;
        if (workspaceFolders && workspaceFolders.length > 0) {
            const wsPath = workspaceFolders[0].uri.fsPath;
            const candidates = [
                path.join(wsPath, 'jabline.exe'),
                path.join(wsPath, 'jabline'),
            ];
            for (const candidate of candidates) {
                const fs = require('fs');
                if (fs.existsSync(candidate)) {
                    command = candidate;
                    break;
                }
            }
        }
    }

    // 3. Fall back to PATH
    if (!command) {
        command = process.platform === 'win32' ? 'jabline.exe' : 'jabline';
    }

    console.log('[Jabline] Using LSP executable:', command);

    const serverOptions = {
        run:   { command: command, args: ['lsp'], transport: TransportKind.stdio },
        debug: { command: command, args: ['lsp'], transport: TransportKind.stdio }
    };

    const clientOptions = {
        documentSelector: [{ scheme: 'file', language: 'jabline' }],
        synchronize: {
            fileEvents: workspace.createFileSystemWatcher('**/*.jb')
        },
        // Trace LSP communication — set to 'verbose' to debug
        traceOutputChannel: window.createOutputChannel('Jabline LSP Trace')
    };

    client = new LanguageClient(
        'jablineLanguageServer',
        'Jabline Language Server',
        serverOptions,
        clientOptions
    );

    client.start();
}

function deactivate() {
    if (!client) {
        return undefined;
    }
    return client.stop();
}

module.exports = { activate, deactivate };

