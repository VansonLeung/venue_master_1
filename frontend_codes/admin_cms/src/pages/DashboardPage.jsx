import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { bookingService } from '@/services/booking.service'
import { venueService } from '@/services/venue.service'
import { facilityService } from '@/services/facility.service'
import { userService } from '@/services/user.service'
import { Building, Warehouse, Calendar, Users } from 'lucide-react'
import { useToast } from '@/components/ui/use-toast'

export default function DashboardPage() {
  const [stats, setStats] = useState({
    venues: 0,
    facilities: 0,
    bookings: 0,
    users: 0,
  })
  const [loading, setLoading] = useState(true)
  const { toast } = useToast()

  useEffect(() => {
    fetchStats()
  }, [])

  const fetchStats = async () => {
    try {
      const [venues, facilities, bookings, users] = await Promise.all([
        venueService.getVenues({ limit: 1 }),
        facilityService.getFacilities({ limit: 1 }),
        bookingService.getBookings({ limit: 1 }),
        userService.getUsers({ limit: 1 }),
      ])

      setStats({
        venues: venues.length || 0,
        facilities: facilities.length || 0,
        bookings: bookings.length || 0,
        users: users.length || 0,
      })
    } catch (error) {
      console.error('Error fetching stats:', error)
      toast({
        title: 'Error',
        description: 'Failed to fetch dashboard statistics',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  const statCards = [
    {
      title: 'Total Venues',
      value: stats.venues,
      icon: Building,
      color: 'text-blue-600',
      bgColor: 'bg-blue-100',
    },
    {
      title: 'Total Facilities',
      value: stats.facilities,
      icon: Warehouse,
      color: 'text-green-600',
      bgColor: 'bg-green-100',
    },
    {
      title: 'Total Bookings',
      value: stats.bookings,
      icon: Calendar,
      color: 'text-purple-600',
      bgColor: 'bg-purple-100',
    },
    {
      title: 'Total Users',
      value: stats.users,
      icon: Users,
      color: 'text-orange-600',
      bgColor: 'bg-orange-100',
    },
  ]

  if (loading) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {[1, 2, 3, 4].map((i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader className="pb-2">
                <div className="h-4 bg-gray-200 rounded w-24"></div>
              </CardHeader>
              <CardContent>
                <div className="h-8 bg-gray-200 rounded w-16"></div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">
          Welcome to Venue Master Admin Panel
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat) => {
          const Icon = stat.icon
          return (
            <Card key={stat.title}>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {stat.title}
                </CardTitle>
                <div className={`p-2 rounded-lg ${stat.bgColor}`}>
                  <Icon className={`h-5 w-5 ${stat.color}`} />
                </div>
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold">{stat.value}</div>
              </CardContent>
            </Card>
          )
        })}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Activity tracking coming soon...
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <p className="text-sm text-muted-foreground">
              Use the sidebar to navigate to different sections
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
