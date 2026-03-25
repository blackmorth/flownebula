import { describe, expect, it } from 'vitest'
import buildTree from './buildTree'

describe('buildTree', () => {
  it('builds a recursive tree from payload nodes and edges', () => {
    const payload = {
      root: 'n1',
      nodes: {
        n1: { nodeId: 'root', inclusive_cost: { wt: 20 } },
        n2: { nodeId: 'childA', inclusive_cost: { wt: 15 } },
        n3: { nodeId: 'childB', inclusive_cost: { wt: 5 } },
      },
      edges: {
        e1: { edgeId: 'e1', caller: 'n1', callee: 'n2', cost: { wt: 15 } },
        e2: { edgeId: 'e2', caller: 'n2', callee: 'n3', cost: { wt: 5 } },
      },
    }

    expect(buildTree(payload)).toEqual({
      id: 'n1',
      name: 'root',
      cost: 20,
      children: [
        {
          id: 'n2#e1',
          name: 'childA',
          cost: 15,
          children: [
            {
              id: 'n3#e2',
              name: 'childB',
              cost: 5,
              children: [],
            },
          ],
        },
      ],
    })
  })

  it('uses edge wall time for recursion entries instead of forcing 0', () => {
    const payload = {
      root: 'n1',
      nodes: {
        n1: { nodeId: 'root', inclusive_cost: { wt: 20 } },
        n2: { nodeId: 'childA', inclusive_cost: { wt: 15 } },
      },
      edges: {
        e1: { edgeId: 'e1', caller: 'n1', callee: 'n2', cost: { wt: 15 } },
        e2: { edgeId: 'e2', caller: 'n2', callee: 'n2', cost: { wt: 9 } },
      },
    }

    expect(buildTree(payload)).toEqual({
      id: 'n1',
      name: 'root',
      cost: 20,
      children: [
        {
          id: 'n2#e1',
          name: 'childA',
          cost: 15,
          children: [
            {
              id: 'n2#e2',
              name: 'childA (recursion)',
              cost: 9,
              children: [],
            },
          ],
        },
      ],
    })
  })
})
