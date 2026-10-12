import { apiClient } from './client'

export interface ImageStudioGalleryEntry {
  id: string
  url: string
  prompt: string
  model: string
  size: string
  format: string
  created_at: number
  expires_at: number
}

export interface SaveImageStudioGalleryRequest {
  image_data_url?: string
  image_url?: string
  prompt: string
  model: string
  size: string
  format: string
}

export async function listImageStudioGallery(): Promise<ImageStudioGalleryEntry[]> {
  const { data } = await apiClient.get<ImageStudioGalleryEntry[]>('/image-studio/gallery')
  return data
}

export async function saveImageStudioGallery(
  request: SaveImageStudioGalleryRequest,
): Promise<ImageStudioGalleryEntry> {
  // Saving includes uploading the image to object storage, which can take several minutes.
  const { data } = await apiClient.post<ImageStudioGalleryEntry>('/image-studio/gallery', request, {
    timeout: 5 * 60 * 1000,
  })
  return data
}

export const imageStudioGalleryAPI = {
  list: listImageStudioGallery,
  save: saveImageStudioGallery,
}

export default imageStudioGalleryAPI
