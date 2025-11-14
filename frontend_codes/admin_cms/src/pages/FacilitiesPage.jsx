import { useEffect, useState } from 'react'
import { facilityService } from '@/services/facility.service'
import { venueService } from '@/services/venue.service'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useToast } from '@/components/ui/use-toast'
import { Plus, Edit, Trash2 } from 'lucide-react'
import { formatCurrency } from '@/lib/utils'

export default function FacilitiesPage() {
  const [facilities, setFacilities] = useState([])
  const [venues, setVenues] = useState([])
  const [loading, setLoading] = useState(true)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingFacility, setEditingFacility] = useState(null)
  const [formData, setFormData] = useState({
    venueId: '',
    name: '',
    description: '',
    surface: '',
    openAt: '08:00',
    closeAt: '22:00',
    available: true,
    weekdayRateCents: '',
    weekendRateCents: '',
    currency: 'USD',
  })
  const { toast } = useToast()

  useEffect(() => {
    fetchData()
  }, [])

  const fetchData = async () => {
    try {
      const [facilitiesData, venuesData] = await Promise.all([
        facilityService.getFacilities({ limit: 100 }),
        venueService.getVenues({ limit: 100 }),
      ])
      setFacilities(facilitiesData)
      setVenues(venuesData)
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch data',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)

    const payload = {
      ...formData,
      weekdayRateCents: formData.weekdayRateCents ? parseInt(formData.weekdayRateCents) : null,
      weekendRateCents: formData.weekendRateCents ? parseInt(formData.weekendRateCents) : null,
    }

    try {
      if (editingFacility) {
        await facilityService.updateFacility(editingFacility.id, payload)
        toast({
          title: 'Success',
          description: 'Facility updated successfully',
        })
      } else {
        await facilityService.createFacility(payload)
        toast({
          title: 'Success',
          description: 'Facility created successfully',
        })
      }
      setDialogOpen(false)
      resetForm()
      fetchData()
    } catch (error) {
      toast({
        title: 'Error',
        description: error.response?.data?.message || 'Operation failed',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  const handleEdit = (facility) => {
    setEditingFacility(facility)
    setFormData({
      venueId: facility.venueId,
      name: facility.name,
      description: facility.description,
      surface: facility.surface,
      openAt: facility.openAt,
      closeAt: facility.closeAt,
      available: facility.available,
      weekdayRateCents: facility.weekdayRateCents || '',
      weekendRateCents: facility.weekendRateCents || '',
      currency: facility.currency,
    })
    setDialogOpen(true)
  }

  const handleDelete = async (id) => {
    if (!confirm('Are you sure you want to delete this facility?')) return

    try {
      await facilityService.deleteFacility(id)
      toast({
        title: 'Success',
        description: 'Facility deleted successfully',
      })
      fetchData()
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete facility',
        variant: 'destructive',
      })
    }
  }

  const resetForm = () => {
    setFormData({
      venueId: '',
      name: '',
      description: '',
      surface: '',
      openAt: '08:00',
      closeAt: '22:00',
      available: true,
      weekdayRateCents: '',
      weekendRateCents: '',
      currency: 'USD',
    })
    setEditingFacility(null)
  }

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target
    setFormData({
      ...formData,
      [name]: type === 'checkbox' ? checked : value,
    })
  }

  if (loading && facilities.length === 0) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading facilities...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Facilities</h1>
          <p className="text-muted-foreground">Manage facility resources</p>
        </div>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button onClick={resetForm}>
              <Plus className="mr-2 h-4 w-4" />
              Add Facility
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>
                {editingFacility ? 'Edit Facility' : 'Create New Facility'}
              </DialogTitle>
              <DialogDescription>
                {editingFacility
                  ? 'Update facility information'
                  : 'Add a new facility to the system'}
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={handleSubmit}>
              <div className="grid gap-4 py-4">
                <div className="grid gap-2">
                  <Label htmlFor="venueId">Venue</Label>
                  <Select
                    name="venueId"
                    value={formData.venueId}
                    onValueChange={(value) =>
                      setFormData({ ...formData, venueId: value })
                    }
                    required
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select a venue" />
                    </SelectTrigger>
                    <SelectContent>
                      {venues.map((venue) => (
                        <SelectItem key={venue.id} value={venue.id}>
                          {venue.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="name">Facility Name</Label>
                  <Input
                    id="name"
                    name="name"
                    value={formData.name}
                    onChange={handleChange}
                    required
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="description">Description</Label>
                  <Input
                    id="description"
                    name="description"
                    value={formData.description}
                    onChange={handleChange}
                    required
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="surface">Surface Type</Label>
                  <Input
                    id="surface"
                    name="surface"
                    placeholder="e.g., Grass, Clay, Hard Court"
                    value={formData.surface}
                    onChange={handleChange}
                    required
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="grid gap-2">
                    <Label htmlFor="openAt">Opening Time</Label>
                    <Input
                      id="openAt"
                      name="openAt"
                      type="time"
                      value={formData.openAt}
                      onChange={handleChange}
                      required
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="closeAt">Closing Time</Label>
                    <Input
                      id="closeAt"
                      name="closeAt"
                      type="time"
                      value={formData.closeAt}
                      onChange={handleChange}
                      required
                    />
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="grid gap-2">
                    <Label htmlFor="weekdayRateCents">Weekday Rate (cents)</Label>
                    <Input
                      id="weekdayRateCents"
                      name="weekdayRateCents"
                      type="number"
                      placeholder="5000 = $50.00"
                      value={formData.weekdayRateCents}
                      onChange={handleChange}
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="weekendRateCents">Weekend Rate (cents)</Label>
                    <Input
                      id="weekendRateCents"
                      name="weekendRateCents"
                      type="number"
                      placeholder="7500 = $75.00"
                      value={formData.weekendRateCents}
                      onChange={handleChange}
                    />
                  </div>
                </div>
                <div className="flex items-center space-x-2">
                  <Switch
                    id="available"
                    checked={formData.available}
                    onCheckedChange={(checked) =>
                      setFormData({ ...formData, available: checked })
                    }
                  />
                  <Label htmlFor="available">Available for booking</Label>
                </div>
              </div>
              <DialogFooter>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    setDialogOpen(false)
                    resetForm()
                  }}
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={loading}>
                  {editingFacility ? 'Update' : 'Create'}
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Facilities</CardTitle>
          <CardDescription>
            A list of all facilities in the system
          </CardDescription>
        </CardHeader>
        <CardContent>
          {facilities.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-muted-foreground">No facilities found</p>
              <p className="text-sm text-muted-foreground mt-1">
                Click "Add Facility" to create your first facility
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Surface</TableHead>
                  <TableHead>Hours</TableHead>
                  <TableHead>Rate</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {facilities.map((facility) => (
                  <TableRow key={facility.id}>
                    <TableCell className="font-medium">
                      {facility.name}
                    </TableCell>
                    <TableCell>{facility.surface}</TableCell>
                    <TableCell>
                      {facility.openAt} - {facility.closeAt}
                    </TableCell>
                    <TableCell>
                      {facility.weekdayRateCents
                        ? formatCurrency(facility.weekdayRateCents, facility.currency)
                        : 'N/A'}
                    </TableCell>
                    <TableCell>
                      <span
                        className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                          facility.available
                            ? 'bg-green-100 text-green-800'
                            : 'bg-red-100 text-red-800'
                        }`}
                      >
                        {facility.available ? 'Available' : 'Unavailable'}
                      </span>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleEdit(facility)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleDelete(facility.id)}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
