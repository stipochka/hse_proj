import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter as Router } from 'react-router-dom'
import { ConfigProvider, Spin } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import App from './App'
import { initKeycloak } from '@/shared/lib/keycloak'
import { useAuthStore } from '@/app/store/authStore'
import './index.css'

const root = ReactDOM.createRoot(document.getElementById('root')!)

root.render(
  <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
    <Spin size="large" tip="Авторизация..." />
  </div>
)

initKeycloak()
  .then((user) => {
    if (user) {
      useAuthStore.getState().setUser(user)
    }
    root.render(
      <React.StrictMode>
        <Router>
          <ConfigProvider locale={ruRU}>
            <App />
          </ConfigProvider>
        </Router>
      </React.StrictMode>
    )
  })
  .catch((error) => {
    console.error('Keycloak init failed:', error)
    root.render(
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <h1>Ошибка авторизации</h1>
        <p>Не удалось инициализировать систему аутентификации. Попробуйте обновить страницу.</p>
      </div>
    )
  })
