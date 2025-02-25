import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.jsx'
import {BrowserRouter} from "react-router-dom";
import {HeroUIProvider} from "@heroui/react";

createRoot(document.getElementById('root')).render(
  <StrictMode>
      <HeroUIProvider>
          <BrowserRouter>
              <App />
          </BrowserRouter>
      </HeroUIProvider>
  </StrictMode>,
)
