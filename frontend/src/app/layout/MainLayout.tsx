import { useState } from 'react'
import { Layout, Menu, Button, Dropdown, Space, Avatar, Typography } from 'antd'
import type { MenuProps } from 'antd'
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  LogoutOutlined,
  DashboardOutlined,
  PlusCircleOutlined,
  FileTextOutlined,
  AuditOutlined,
  TeamOutlined,
  ExportOutlined,
  DownOutlined,
} from '@ant-design/icons'
import { Link, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/app/store/authStore'
import { logout as keycloakLogout } from '@/shared/lib/keycloak'

const { Header, Sider, Content } = Layout
const { Text } = Typography

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
    { key: '/', label: <Link to="/">Дашборд</Link>, icon: <DashboardOutlined /> },
    { key: '/submit', label: <Link to="/submit">Подать активность</Link>, icon: <PlusCircleOutlined /> },
    { key: '/activities', label: <Link to="/activities">Мои активности</Link>, icon: <FileTextOutlined /> },
    { key: '/export', label: <Link to="/export">Экспорт</Link>, icon: <ExportOutlined /> },
  ]

  const adminItems = [
    { key: '/', label: <Link to="/">Дашборд</Link>, icon: <DashboardOutlined /> },
    { key: '/activities', label: <Link to="/activities">На проверку</Link>, icon: <AuditOutlined /> },
    { key: '/group-students', label: <Link to="/group-students">Студенты</Link>, icon: <TeamOutlined /> },
    { key: '/export', label: <Link to="/export">Экспорт</Link>, icon: <ExportOutlined /> },
  ]

  const userDropdown: MenuProps['items'] = [
    { key: 'logout', label: 'Выйти', icon: <LogoutOutlined />, onClick: handleLogout },
  ]

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        width={220}
        style={{ background: '#fff', borderRight: '1px solid #f0f0f0' }}
      >
        <div style={{
          height: 64,
          display: 'flex',
          alignItems: 'center',
          justifyContent: collapsed ? 'center' : 'flex-start',
          padding: collapsed ? 0 : '0 20px',
          borderBottom: '1px solid #f0f0f0',
          overflow: 'hidden',
        }}>
          {collapsed ? (
            <div style={{
              width: 32, height: 32, borderRadius: 8,
              background: '#1677ff', display: 'flex',
              alignItems: 'center', justifyContent: 'center',
            }}>
              <Text style={{ color: '#fff', fontWeight: 700, fontSize: 14 }}>H</Text>
            </div>
          ) : (
            <Space>
              <div style={{
                width: 32, height: 32, borderRadius: 8,
                background: '#1677ff', display: 'flex',
                alignItems: 'center', justifyContent: 'center', flexShrink: 0,
              }}>
                <Text style={{ color: '#fff', fontWeight: 700, fontSize: 14 }}>H</Text>
              </div>
              <Text strong style={{ fontSize: 15 }}>HSE Активности</Text>
            </Space>
          )}
        </div>
        <Menu
          mode="inline"
          selectedKeys={[location.pathname]}
          items={admin ? adminItems : studentItems}
          style={{ border: 'none', marginTop: 8 }}
        />
      </Sider>

      <Layout>
        <Header style={{
          background: '#fff',
          borderBottom: '1px solid #f0f0f0',
          padding: '0 24px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          height: 64,
        }}>
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
            style={{ fontSize: 16, width: 40, height: 40 }}
          />
          {user && (
            <Dropdown menu={{ items: userDropdown }} trigger={['click']}>
              <Space style={{ cursor: 'pointer' }}>
                <Avatar
                  size={32}
                  style={{ background: '#1677ff', fontSize: 13 }}
                >
                  {(user.firstName || user.username || '?')[0].toUpperCase()}
                </Avatar>
                <Text style={{ fontSize: 14 }}>{user.firstName || user.username}</Text>
                <DownOutlined style={{ fontSize: 11, color: '#8c8c8c' }} />
              </Space>
            </Dropdown>
          )}
        </Header>

        <Content style={{ margin: '24px', minHeight: 'calc(100vh - 112px)' }}>
          {children}
        </Content>
      </Layout>
    </Layout>
  )
}

export default MainLayout
