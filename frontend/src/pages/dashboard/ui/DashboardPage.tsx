import { useEffect, useState } from 'react'
import { Row, Col, Card, Statistic, Table, Spin, Empty, Tabs, Tag, Typography } from 'antd'
import {
  FileTextOutlined,
  TrophyOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import FilterBar from '@/shared/components/FilterBar'
import { dashboardAPI } from '@/shared/api/dashboard'
import { useAuthStore } from '@/app/store/authStore'
import type { FilterOptions, DashboardMe, StudentStats } from '@/shared/types'

const { Title } = Typography

const isAdmin = (roles: string[]) =>
  roles.includes('group_admin') || roles.includes('super_admin')

const statusColor: Record<string, string> = {
  SUBMITTED: 'processing',
  EVALUATED: 'success',
  REJECTED: 'error',
  PENDING: 'default',
}

const statusLabel: Record<string, string> = {
  SUBMITTED: 'На проверке',
  EVALUATED: 'Проверено',
  REJECTED: 'Отклонено',
  PENDING: 'Ожидание',
}

const DashboardPage = () => {
  const user = useAuthStore((state) => state.user)
  const admin = isAdmin(user?.roles ?? [])

  const [loading, setLoading] = useState(true)
  const [myDashboard, setMyDashboard] = useState<DashboardMe | null>(null)
  const [summaryData, setSummaryData] = useState<StudentStats[]>([])

  const fetchDashboardData = async (appliedFilters: FilterOptions = {}) => {
    setLoading(true)
    try {
      if (admin) {
        const [myData, summary] = await Promise.all([
          dashboardAPI.getMyDashboard(),
          dashboardAPI.getSummary(appliedFilters),
        ])
        setMyDashboard(myData)
        setSummaryData(summary)
      } else {
        const myData = await dashboardAPI.getMyDashboard()
        setMyDashboard(myData)
      }
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchDashboardData() }, [])

  const summaryColumns: ColumnsType<StudentStats> = [
    { title: '№', key: 'index', width: 56, render: (_, __, i) => i + 1 },
    {
      title: 'Студент',
      key: 'student',
      ellipsis: true,
      render: (_: unknown, r: StudentStats) => r.student_name || r.student_id,
    },
    { title: 'Группа', dataIndex: 'student_group', key: 'student_group', width: 130 },
    {
      title: 'Активностей',
      dataIndex: 'activity_count',
      key: 'activity_count',
      width: 130,
      sorter: (a, b) => b.activity_count - a.activity_count,
    },
    {
      title: 'Проверено',
      dataIndex: 'evaluated_count',
      key: 'evaluated_count',
      width: 120,
      sorter: (a, b) => b.evaluated_count - a.evaluated_count,
    },
    {
      title: 'Баллы',
      dataIndex: 'total_points',
      key: 'total_points',
      width: 100,
      sorter: (a, b) => b.total_points - a.total_points,
      render: (v) => <strong>{v}</strong>,
    },
    {
      title: 'З.е.',
      dataIndex: 'total_credits',
      key: 'total_credits',
      width: 90,
      sorter: (a, b) => b.total_credits - a.total_credits,
      render: (v) => v.toFixed(2),
    },
  ]

  const myTab = (
    <div>
      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: '80px 0' }}>
          <Spin size="large" />
        </div>
      ) : myDashboard ? (
        <>
          <Row gutter={[20, 20]}>
            <Col xs={24} sm={12} xl={6}>
              <Card bordered={false} style={{ boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}>
                <Statistic
                  title="Всего активностей"
                  value={myDashboard.activity_count}
                  prefix={<FileTextOutlined style={{ color: '#1677ff' }} />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} xl={6}>
              <Card bordered={false} style={{ boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}>
                <Statistic
                  title="Баллы"
                  value={myDashboard.total_points}
                  prefix={<TrophyOutlined style={{ color: '#faad14' }} />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} xl={6}>
              <Card bordered={false} style={{ boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}>
                <Statistic
                  title="Зачётные единицы"
                  value={myDashboard.total_credits.toFixed(2)}
                  prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} xl={6}>
              <Card bordered={false} style={{ boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}>
                <Statistic
                  title="На проверке"
                  value={myDashboard.by_status?.SUBMITTED ?? 0}
                  prefix={<ClockCircleOutlined style={{ color: '#722ed1' }} />}
                />
              </Card>
            </Col>
          </Row>

          {Object.keys(myDashboard.by_status ?? {}).length > 0 && (
            <Row gutter={[20, 20]} style={{ marginTop: 20 }}>
              <Col xs={24} lg={12}>
                <Card
                  title="По статусам"
                  bordered={false}
                  style={{ boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}
                >
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: 12 }}>
                    {Object.entries(myDashboard.by_status).map(([status, count]) => (
                      <div key={status} style={{ minWidth: 120 }}>
                        <div style={{ marginBottom: 4 }}>
                          <Tag color={statusColor[status] ?? 'default'}>
                            {statusLabel[status] ?? status}
                          </Tag>
                        </div>
                        <div style={{ fontSize: 22, fontWeight: 600 }}>{count}</div>
                      </div>
                    ))}
                  </div>
                </Card>
              </Col>
              {Object.keys(myDashboard.by_category ?? {}).length > 0 && (
                <Col xs={24} lg={12}>
                  <Card
                    title="По категориям"
                    bordered={false}
                    style={{ boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}
                  >
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: 12 }}>
                      {Object.entries(myDashboard.by_category).map(([cat, count]) => (
                        <div key={cat} style={{ minWidth: 120 }}>
                          <div style={{ marginBottom: 4, color: '#595959', fontSize: 13 }}>{cat}</div>
                          <div style={{ fontSize: 22, fontWeight: 600 }}>{count}</div>
                        </div>
                      ))}
                    </div>
                  </Card>
                </Col>
              )}
            </Row>
          )}
        </>
      ) : (
        <Empty description="Нет данных" style={{ padding: '60px 0' }} />
      )}
    </div>
  )

  return (
    <div>
      <Title level={3} style={{ marginBottom: 24, fontWeight: 600 }}>Дашборд</Title>

      {admin ? (
        <Tabs
          defaultActiveKey="my"
          items={[
            { key: 'my', label: 'Мои активности', children: myTab },
            {
              key: 'admin',
              label: 'Сводка по студентам',
              children: (
                <div>
                  <FilterBar onFilterChange={fetchDashboardData} />
                  {loading ? (
                    <div style={{ display: 'flex', justifyContent: 'center', padding: '80px 0' }}>
                      <Spin size="large" />
                    </div>
                  ) : summaryData.length === 0 ? (
                    <Empty description="Нет данных" style={{ padding: '60px 0' }} />
                  ) : (
                    <Card bordered={false} style={{ boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}>
                      <Table
                        columns={summaryColumns}
                        dataSource={summaryData}
                        rowKey="student_id"
                        pagination={false}
                        size="middle"
                      />
                    </Card>
                  )}
                </div>
              ),
            },
          ]}
        />
      ) : myTab}
    </div>
  )
}

export default DashboardPage
