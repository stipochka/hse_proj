import { useEffect, useState } from 'react'
import { Row, Col, Card, Statistic, Table, Spin, Empty, Tabs } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import FilterBar from '@/shared/components/FilterBar'
import { dashboardAPI } from '@/shared/api/dashboard'
import { useAuthStore } from '@/app/store/authStore'
import type { FilterOptions, DashboardMe, StudentStats } from '@/shared/types'

const isAdmin = (roles: string[]) =>
  roles.includes('group_admin') || roles.includes('super_admin')

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

  useEffect(() => {
    fetchDashboardData()
  }, [])

  const summaryColumns: ColumnsType<StudentStats> = [
    {
      title: '№',
      key: 'index',
      width: 50,
      render: (_, __, index) => index + 1,
    },
    {
      title: 'ID Студента',
      dataIndex: 'student_id',
      key: 'student_id',
    },
    {
      title: 'Группа',
      dataIndex: 'student_group',
      key: 'student_group',
    },
    {
      title: 'Активностей',
      dataIndex: 'activity_count',
      key: 'activity_count',
      sorter: (a, b) => b.activity_count - a.activity_count,
    },
    {
      title: 'Проверено',
      dataIndex: 'evaluated_count',
      key: 'evaluated_count',
      sorter: (a, b) => b.evaluated_count - a.evaluated_count,
    },
    {
      title: 'Баллы',
      dataIndex: 'total_points',
      key: 'total_points',
      sorter: (a, b) => b.total_points - a.total_points,
    },
    {
      title: 'З.е.',
      dataIndex: 'total_credits',
      key: 'total_credits',
      sorter: (a, b) => b.total_credits - a.total_credits,
      render: (value) => value.toFixed(2),
    },
  ]

  const myTab = (
    <div>
      {myDashboard ? (
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic title="Активностей" value={myDashboard.activity_count} prefix="📋" />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic title="Баллы" value={myDashboard.total_points} prefix="📊" />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Зачетные единицы"
                value={myDashboard.total_credits.toFixed(2)}
                prefix="✓"
              />
            </Card>
          </Col>
        </Row>
      ) : (
        <Empty description="Нет данных" />
      )}

      {myDashboard && myDashboard.by_status && Object.keys(myDashboard.by_status).length > 0 && (
        <Row gutter={[16, 16]} style={{ marginTop: '24px' }}>
          <Col xs={24} lg={12}>
            <Card title="По статусам">
              <Row gutter={[16, 16]}>
                {Object.entries(myDashboard.by_status).map(([status, count]) => (
                  <Col xs={12} sm={8} key={status}>
                    <Statistic title={status} value={count} />
                  </Col>
                ))}
              </Row>
            </Card>
          </Col>
          <Col xs={24} lg={12}>
            <Card title="По категориям">
              <Row gutter={[16, 16]}>
                {Object.entries(myDashboard.by_category).map(([category, count]) => (
                  <Col xs={12} sm={8} key={category}>
                    <Statistic title={category} value={count} />
                  </Col>
                ))}
              </Row>
            </Card>
          </Col>
        </Row>
      )}
    </div>
  )

  return (
    <div>
      <h1 className="text-3xl font-bold mb-6">Дашборд</h1>

      {loading ? (
        <div className="flex justify-center items-center h-96">
          <Spin size="large" />
        </div>
      ) : admin ? (
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
                  {summaryData.length === 0 ? (
                    <Empty description="Нет данных" />
                  ) : (
                    <Table
                      columns={summaryColumns}
                      dataSource={summaryData}
                      rowKey={(record) => record.student_id}
                      pagination={false}
                      className="bg-white rounded-lg"
                    />
                  )}
                </div>
              ),
            },
          ]}
        />
      ) : (
        myTab
      )}
    </div>
  )
}

export default DashboardPage
