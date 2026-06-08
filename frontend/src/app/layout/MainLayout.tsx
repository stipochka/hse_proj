import { useState } from 'react'
import { Layout, Menu, Button, Dropdown, Space, Avatar } from 'antd'
import type { MenuProps } from 'antd'

type UserMenuItems = MenuProps['items']
import { MenuFoldOutlined, MenuUnfoldOutlined, LogoutOutlined, UserOutlined } from '@ant-design/icons'
import { Link, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/app/store/authStore'
import { logout as keycloakLogout } from '@/shared/lib/keycloak'

const { Header, Sider, Content } = Layout

interface MainLayoutProps {
  children: React.ReactNode
}

const isAdmin = (roles: string[]) =>
  roles.includes('group_admin') || roles.includes('super_admin')

const MainLayout = ({ children }: MainLayoutProps) => {
  const [collapsed, setCollapsed] = useState(false)
  const location = useLocation()
  const user = useAuthStore((state) => state.user)
  const logoutUser = useAuthStore((state) => state.logout)
  const admin = isAdmin(user?.roles ?? [])

  const handleLogout = () => {
    logoutUser()
    keycloakLogout()
  }

  const studentItems = [
    {
      key: '/',
      label: <Link to="/">Дашборд</Link>,
      icon: <span>📊</span>,
    },
    {
      key: '/submit',
      label: <Link to="/submit">Подать активность</Link>,
      icon: <span>📤</span>,
    },
    {
      key: '/activities',
      label: <Link to="/activities">Мои активности</Link>,
      icon: <span>📋</span>,
    },
    {
      key: '/export',
      label: <Link to="/export">Экспорт</Link>,
      icon: <span>📥</span>,
    },
  ]

  const adminItems = [
    {
      key: '/',
      label: <Link to="/">Дашборд</Link>,
      icon: <span>📊</span>,
    },
    {
      key: '/activities',
      label: <Link to="/activities">Активности на проверку</Link>,
      icon: <span>✓</span>,
    },
    {
      key: '/group-students',
      label: <Link to="/group-students">Студенты группы</Link>,
      icon: <span>👥</span>,
    },
    {
      key: '/export',
      label: <Link to="/export">Экспорт</Link>,
      icon: <span>📥</span>,
    },
  ]

  const menuItems = admin ? adminItems : studentItems

  const userMenuItems: UserMenuItems = user ? [
    {
      key: 'profile',
      label: 'Профиль',
      icon: <UserOutlined />,
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      label: 'Выход',
      icon: <LogoutOutlined />,
      onClick: handleLogout,
    },
  ] : []

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider trigger={null} collapsible collapsed={collapsed} theme="dark">
        <div className="logo p-4 text-white text-lg font-bold text-center">
          {!collapsed && 'HSE Активности'}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
        />
      </Sider>
      <Layout>
        <Header
          className="bg-white shadow-sm flex items-center justify-between"
          style={{ padding: '0 24px' }}
        >
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
            size="large"
          />
          <Space>
            {user ? (
              <Dropdown menu={{ items: userMenuItems }} trigger={['click']}>
                <Space style={{ cursor: 'pointer' }}>
                  <Avatar icon={<UserOutlined />} />
                  <span>{user.firstName || user.username}</span>
                </Space>
              </Dropdown>
            ) : null}
          </Space>
        </Header>
        <Content className="m-6 p-6 bg-gray-50 rounded-lg min-h-[calc(100vh-113px)]">
          {children}
        </Content>
      </Layout>
    </Layout>
  )
}

export default MainLayout
