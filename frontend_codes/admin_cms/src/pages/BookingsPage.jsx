import { useEffect, useState } from 'react'
import { bookingService } from '@/services/booking.service'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useToast } from '@/components/ui/use-toast'
import { formatCurrency, formatDateTime } from '@/lib/utils'

const STATUS_OPTIONS = [
  'PENDING_PAYMENT',
  'CONFIRMED',
  'CANCELLED',
  'COMPLETED',
]

const STATUS_COLORS = {
  PENDING_PAYMENT: 'bg-yellow-100 text-yellow-800',
  CONFIRMED: 'bg-green-100 text-green-800',
  CANCELLED: 'bg-red-100 text-red-800',
  COMPLETED: 'bg-blue-100 text-blue-800',
}

export default function BookingsPage() {
  const [bookings, setBookings] = useState([])
  const [loading, setLoading] = useState(true)
  const { toast } = useToast()

  useEffect(() => {
    fetchBookings()
  }, [])

  const fetchBookings = async () => {
    try {
      const data = await bookingService.getBookings({ limit: 100 })
      setBookings(data)
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch bookings',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  const handleStatusChange = async (bookingId, newStatus) => {
    try {
      await bookingService.updateBookingStatus(bookingId, newStatus)
      toast({
        title: 'Success',
        description: 'Booking status updated successfully',
      })
      fetchBookings()
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to update booking status',
        variant: 'destructive',
      })
    }
  }

  const handleConfirm = async (bookingId) => {
    try {
      await bookingService.confirmBooking(bookingId)
      toast({
        title: 'Success',
        description: 'Booking confirmed successfully',
      })
      fetchBookings()
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to confirm booking',
        variant: 'destructive',
      })
    }
  }

  const handleCancel = async (bookingId) => {
    if (!confirm('Are you sure you want to cancel this booking?')) return

    try {
      await bookingService.cancelBooking(bookingId)
      toast({
        title: 'Success',
        description: 'Booking cancelled successfully',
      })
      fetchBookings()
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to cancel booking',
        variant: 'destructive',
      })
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading bookings...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Bookings</h1>
        <p className="text-muted-foreground">Manage facility bookings</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Bookings</CardTitle>
          <CardDescription>
            A list of all bookings in the system
          </CardDescription>
        </CardHeader>
        <CardContent>
          {bookings.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-muted-foreground">No bookings found</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Booking ID</TableHead>
                    <TableHead>Facility</TableHead>
                    <TableHead>Start Time</TableHead>
                    <TableHead>End Time</TableHead>
                    <TableHead>Amount</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {bookings.map((booking) => (
                    <TableRow key={booking.id}>
                      <TableCell className="font-mono text-xs">
                        {booking.id.substring(0, 8)}...
                      </TableCell>
                      <TableCell>
                        {booking.facility?.name || 'N/A'}
                      </TableCell>
                      <TableCell>{formatDateTime(booking.startsAt)}</TableCell>
                      <TableCell>{formatDateTime(booking.endsAt)}</TableCell>
                      <TableCell className="font-medium">
                        {formatCurrency(booking.amountCents, booking.currency)}
                      </TableCell>
                      <TableCell>
                        <span
                          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            STATUS_COLORS[booking.status] || 'bg-gray-100 text-gray-800'
                          }`}
                        >
                          {booking.status.replace('_', ' ')}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-2">
                          {booking.status === 'PENDING_PAYMENT' && (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleConfirm(booking.id)}
                            >
                              Confirm
                            </Button>
                          )}
                          {(booking.status === 'PENDING_PAYMENT' ||
                            booking.status === 'CONFIRMED') && (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleCancel(booking.id)}
                            >
                              Cancel
                            </Button>
                          )}
                          <Select
                            value={booking.status}
                            onValueChange={(value) =>
                              handleStatusChange(booking.id, value)
                            }
                          >
                            <SelectTrigger className="w-[140px] h-9">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              {STATUS_OPTIONS.map((status) => (
                                <SelectItem key={status} value={status}>
                                  {status.replace('_', ' ')}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
