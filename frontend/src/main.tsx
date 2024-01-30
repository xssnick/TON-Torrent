import React from 'react'
import {createRoot} from 'react-dom/client'
import 'bootstrap/dist/css/bootstrap.min.css';
import './app.scss';
import App from './App'
import {TonConnectUIProvider} from "@tonconnect/ui-react";

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <TonConnectUIProvider manifestUrl="https://ton.org/torrent/manifest.json">
            <App/>
        </TonConnectUIProvider>
    </React.StrictMode>
)
